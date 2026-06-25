<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { api, jsonBody } from '../api'
import { auth, refreshMe } from '../auth'
import { draft, applyJobToDraft } from '../playground'
import Icon from '../components/Icon.vue'
import SelectMenu from '../components/SelectMenu.vue'
import MediaLightbox from '../components/MediaLightbox.vue'
import { points, pointsLabel } from '../credits'
import { sortResolutions } from '../utils/format'

const route = useRoute()

// ---- credits: the logged-in user's REAL server-side balance ----
const credits = computed(() => Number(auth.user?.credits || 0))

const allModels = ref([])      // managed-models list
const presets = ref([])        // video family presets

// Seed every form field from the shared draft (module-level) so navigating
// away from the page and coming back keeps the prompt + selected model +
// params. Each local ref then syncs back into the draft on change.
const mode = ref(draft.mode || 'image')
const modelId = ref(draft.modelId || '')
const prompt = ref(draft.prompt || '')
const ratio = ref(draft.ratio || '')
const resolution = ref(draft.resolution || '')
const duration = ref(draft.duration || '')

watch(mode,       (v) => { draft.mode = v })
watch(modelId,    (v) => { draft.modelId = v })
watch(prompt,     (v) => { draft.prompt = v })
watch(ratio,      (v) => { draft.ratio = v })
watch(resolution, (v) => { draft.resolution = v })
watch(duration,   (v) => { draft.duration = v })

const refImages = ref([])      // [{ name, dataUrl }]
const fileInput = ref(null)

const busy = ref(false)
// submitting = run() owns the busy/current state end-to-end while a /generate
// call is in flight. The 2s poll() must NOT touch busy or current during this
// window, or it races run() and the controls flicker unlocked mid-generation.
const submitting = ref(false)
// Gateway-timeout statuses our BACKEND never emits — they mean a CDN/proxy
// (e.g. EdgeOne 524) gave up waiting while the synchronous /generate is STILL
// rendering server-side. Treat these as "still running", NOT a failure: keep the
// controls locked and let poll() follow the live job to completion (出图).
const GATEWAY_TIMEOUT = new Set([0, 408, 504, 520, 521, 522, 523, 524, 525])
const error = ref('')
const statusText = ref('')

// Only ever show the latest generation on the right side. Each new run
// replaces it; the persistent history lives at /logs (UserLogsView).
// `current` is restored from the server on mount and refreshed via /jobs/mine
// polling, so a reload, a parallel tab, or a different browser sees the same
// in-flight job and the same final result without re-running anything.
const current = ref(null)
const lightbox = ref(null)
const toast = ref('')
let pollTimer = null

// ---- derived ----
const models = computed(() =>
  allModels.value.filter((m) => m.enabled !== false && m.type === mode.value),
)
const modelOptions = computed(() =>
  models.value.map((m) => ({ value: m.id, label: m.name || m.id })),
)
const model = computed(() => allModels.value.find((m) => m.id === modelId.value) || null)
const familyPreset = computed(() => {
  if (mode.value !== 'video' || !model.value) return null
  return presets.value.find((p) => p.key === model.value.id) || null
})

const ratios = computed(() => {
  const fromModel = model.value?.ratios || []
  if (fromModel.length) return fromModel
  return mode.value === 'video' ? ['16x9'] : ['1:1']
})
// Firefly Image 5 instruct-edit derives the aspect ratio from the reference
// image — hide the ratio picker (backend also omits aspectRatio) when a ref is
// attached, otherwise the request is rejected with a validation error.
const showRatio = computed(() => !(modelId.value === 'firefly-image-5' && refImages.value.length > 0))
const resolutions = computed(() => {
  const fromModel = model.value?.resolutions || []
  if (fromModel.length) return sortResolutions(fromModel)
  // Legacy record with no declared tiers: fall back to the priced tiers so we
  // never offer (or default to) a resolution the server has no price for.
  const priced = Object.keys(model.value?.prices || {})
  if (priced.length) return sortResolutions(priced)
  return mode.value === 'video' ? ['720p'] : ['1K']
})
const durations = computed(() => {
  // duration_prices arrives as a JSON object whose keys Go sorts alphabetically
  // ("10s" before "5s"). Re-sort by the numeric seconds so the shortest is first.
  const keys = Object.keys(model.value?.duration_prices || {})
    .sort((a, b) => parseFloat(a) - parseFloat(b))
  if (keys.length) return keys
  return familyPreset.value?.durations || ['5s']
})

const maxRefs = computed(() => {
  if (mode.value === 'video') {
    const a = Number(familyPreset.value?.max_reference_images || 0)
    const b = Number(model.value?.max_reference_images || 0)
    return Math.max(a, b)
  }
  // Image-to-image: honor the model's configured max (gpt-image-2=3,
  // seedream-4.5=6, flux-klein-2=4 …). Fall back to 1 when image_to_image is on
  // but no count was set.
  const m = Number(model.value?.max_reference_images || 0)
  if (m > 0) return m
  return model.value?.image_to_image ? 1 : 0
})
const refMode = computed(() => familyPreset.value?.reference_mode || model.value?.reference_mode || 'none')
// Most video models (veo31, luma) support pure text2video, so refs are optional.
// A model can opt into strict image-to-video by declaring `requires_reference`
// in its preset (e.g. runway-gen4-turbo) — then a first-frame image is mandatory.
const refsRequired = computed(() =>
  mode.value === 'video' && !!familyPreset.value?.requires_reference)

// ---- price (per generation, derived from selected model + params) ----
// 代理用户走代理价:某档设了代理价就用它,否则回退普通价(支持的档位始终由普通价决定)。
const isAgent = computed(() => auth.user?.role === 'agent')
function tierPrice(normalMap, agentMap, key) {
  const n = (normalMap || {})[key]
  if (n == null) return null            // 不支持该档(由普通价决定)
  if (isAgent.value) {
    const a = (agentMap || {})[key]
    if (a != null) return Number(a)
  }
  return Number(n)
}
const price = computed(() => {
  if (!model.value) return null
  const m = model.value
  if (mode.value === 'video') {
    const rp = tierPrice(m.prices, m.prices_agent, resolution.value)
    const dp = tierPrice(m.duration_prices, m.duration_prices_agent, duration.value)
    if (rp == null || dp == null) return null
    return rp + dp
  }
  return tierPrice(m.prices, m.prices_agent, resolution.value)
})
const priceLabel = computed(() => price.value == null ? '—' : pointsLabel(price.value))
const canAfford = computed(() => price.value == null || credits.value >= price.value)

// ---- helpers ----
function firstOf(arr) {
  return (arr && arr.length) ? arr[0] : ''
}
// Selecting a model (or switching image/video) resets each picker to that
// model's FIRST tier — the default should always be the first option, not
// whatever was carried over from the previously-selected model.
function applyModelDefaults() {
  ratio.value = firstOf(ratios.value)
  resolution.value = firstOf(resolutions.value)
  duration.value = firstOf(durations.value)
  // 切换模型保留已上传的参考图,只按新模型的上限裁剪(上限为 0 则清空)。
  if (refImages.value.length > maxRefs.value) {
    refImages.value = refImages.value.slice(0, maxRefs.value)
  }
}
function selectModel(id) {
  modelId.value = id
  applyModelDefaults()
}

function setMode(m) {
  if (mode.value === m) return
  mode.value = m
  // pick a default model of the new kind, if any
  const first = allModels.value.find((x) => x.enabled !== false && x.type === m)
  modelId.value = first?.id || ''
  applyModelDefaults()
}

function openPicker() { fileInput.value && fileInput.value.click() }
// Backend rejects reference images over 8MB (maxReferenceImageBytes). Enforce it
// here at pick time so an oversized image fails fast with a clear message instead
// of charging + failing upstream after the upload.
const MAX_REF_BYTES = 8 * 1024 * 1024
function onFiles(ev) {
  const files = Array.from(ev.target.files || [])
  const room = Math.max(0, maxRefs.value - refImages.value.length)
  const tooBig = []
  let added = 0
  for (const f of files) {
    if (added >= room) break
    if (f.size > MAX_REF_BYTES) { tooBig.push(f.name); continue }
    const reader = new FileReader()
    reader.onload = () => refImages.value.push({ name: f.name, dataUrl: reader.result })
    reader.readAsDataURL(f)
    added++
  }
  error.value = tooBig.length
    ? `图片超过 8MB 已跳过：${tooBig.join('、')}（请压缩后再传）`
    : ''
  if (ev.target) ev.target.value = ''
}
function removeRef(i) { refImages.value.splice(i, 1) }

// Re-hydrate reference thumbnails from server URLs (after a reload). Fetches
// each /images URL (same-origin, cookie-authed) and converts to a data URL so
// the thumbnail renders AND the ref can be re-submitted unchanged. Shared by
// image and video — both persist their refs the same way server-side.
// Re-display refs by URL only — the thumbnail renders straight from /images/<ref>.
// We DON'T fetch+convert here; conversion to base64 happens lazily at submit time
// (refToBase64), so a ref that's only viewed never needs a network round-trip.
function restoreRefs(urls) {
  if (!Array.isArray(urls) || !urls.length) return
  if (refImages.value.length) return   // don't clobber refs the user already added
  refImages.value = urls.map((u) => ({ name: 'ref', url: u }))
}

// refToBase64 yields the raw base64 the backend expects, from either a freshly
// uploaded ref (dataUrl) or a restored one (url → fetch). Returns '' on failure.
async function refToBase64(r) {
  try {
    if (r.dataUrl) return r.dataUrl.replace(/^data:[^,]*,/, '')
    if (r.url) {
      const blob = await (await fetch(r.url)).blob()
      const dataUrl = await new Promise((res, rej) => {
        const fr = new FileReader()
        fr.onload = () => res(fr.result)
        fr.onerror = rej
        fr.readAsDataURL(blob)
      })
      return dataUrl.replace(/^data:[^,]*,/, '')
    }
  } catch { /* fall through */ }
  return ''
}

let toastTimer = null
function flash(msg) {
  toast.value = msg
  clearTimeout(toastTimer)
  toastTimer = setTimeout(() => (toast.value = ''), 1800)
}

async function copyLink(url) {
  try {
    const abs = url.startsWith('http') ? url : location.origin + url
    await navigator.clipboard.writeText(abs)
    flash('链接已复制')
  } catch {
    flash('复制失败')
  }
}

// ---- generate ----
async function run() {
  if (!modelId.value) { error.value = '请选择模型'; return }
  if (!prompt.value.trim()) { error.value = '请输入提示词'; return }
  if (refsRequired.value && refImages.value.length < 1) {
    error.value = '该视频模型需要至少 1 张参考图 (首帧)'
    return
  }
  if (price.value == null) {
    error.value = '该参数组合未定价 (留空 = 不支持)'
    return
  }
  if (!canAfford.value) {
    error.value = `积分不足 — 需要 ${pointsLabel(price.value)},余额 ${pointsLabel(credits.value)}`
    return
  }
  const job = {
    id: Math.random().toString(36).slice(2, 10),
    model: modelId.value,
    kind: mode.value,
    prompt: prompt.value,
    ratio: ratio.value,
    resolution: resolution.value,
    duration: mode.value === 'video' ? duration.value : '',
    refs: refImages.value.length,
    status: 'pending',
    url: '',
    error: '',
    elapsed_ms: 0,
    charged: price.value,
    ts: Date.now(),
  }
  current.value = job

  busy.value = true
  submitting.value = true
  error.value = ''
  statusText.value = mode.value === 'video' ? '生成视频中 (约 1–3 分钟)…' : '生成中…'

  try {
    // Optimistically deduct the price from the displayed balance right away. The
    // server debits BEFORE generating (which can take minutes for video), so
    // otherwise 余额 looks unchanged the whole time. The success response carries
    // the authoritative balance (reconciled below); a failure refunds + refreshMe.
    if (auth.user && price.value != null) {
      auth.user.credits = Math.max(0, Number(auth.user.credits || 0) - price.value)
    }

    const payload = {
      model: modelId.value,
      prompt: prompt.value,
      ratio: ratio.value,
      resolution: resolution.value,
    }
    if (mode.value === 'video') payload.duration = duration.value
    if (refImages.value.length) {
      // Backend accepts raw base64 only — convert each ref (uploaded dataUrl or
      // restored /images URL) to base64 at submit time.
      const refs = await Promise.all(refImages.value.map(refToBase64))
      payload.reference_images = refs.filter(Boolean)
    }

    // Single charged call: the server debits the price atomically BEFORE
    // generating and refunds on failure, so the client can't skip the charge.
    const r = await api('/generate', jsonBody('POST', payload))

    if (r.ok && r.data?.url) {
      job.status = 'done'
      job.url = r.data.url
      job.elapsed_ms = r.data.elapsed_ms
      job.charged = r.data.charged ?? price.value
      if (auth.user && r.data.credits != null) auth.user.credits = r.data.credits
      statusText.value = `完成 · 扣费 ${pointsLabel(job.charged)} · ${(r.data.elapsed_ms / 1000).toFixed(1)}s · 余额 ${pointsLabel(credits.value)}`
      busy.value = false                       // 出图 → 解锁
    } else if (GATEWAY_TIMEOUT.has(r.status)) {
      // CDN/代理回源超时(如 EdgeOne 524)—— 后端仍在生成。不当失败、不解锁:
      // 保持 busy=true,交给下面的 poll() + 2s 轮询跟到出图("不出图就不闪")。
      statusText.value = mode.value === 'video' ? '生成视频中 (约 1–3 分钟)…' : '生成中…'
    } else {
      // 真失败:服务端已退款,resync 余额并解锁。
      await refreshMe()
      job.status = 'failed'
      job.error = r.data?.detail || `失败 (${r.status})`
      statusText.value = ''
      busy.value = false                       // 真失败 → 解锁
    }
  } finally {
    // Hand control back to poll(); busy is left as set above (locked when the
    // job is still rendering after a gateway timeout).
    submitting.value = false
  }
  // Sync real server state: poll() picks up the live pending job (replacing our
  // optimistic one with the real id) and will unlock + show the result the
  // moment it finishes — so a 524 mid-flight never leaves the UI unlocked.
  poll()
}

// Recover the current generation from the server: any pending job for this
// user lives in event_log, so reload / parallel tab / parallel browser can
// all see the same in-flight state and the same final result.
async function poll() {
  // While run() is mid-submit it fully owns busy/current — don't race it.
  if (submitting.value) return
  const r = await api('/jobs/mine')
  if (!r.ok) return
  const { pending, latest } = r.data || {}
  if (pending) {
    busy.value = true
    if (!statusText.value) {
      statusText.value = pending.kind === 'video' ? '生成视频中 (约 1–3 分钟)…' : '生成中…'
    }
    if (!current.value || current.value.id !== pending.id) {
      current.value = { ...pending }
      // Replay the pending job's params onto the form so a fresh tab shows
      // what's cooking — and writes them into the cross-component draft.
      applyJobToDraft(pending)
      mode.value = pending.kind === 'video' ? 'video' : 'image'
      modelId.value = pending.model || modelId.value
      prompt.value = pending.prompt || prompt.value
      ratio.value = pending.ratio || ratio.value
      resolution.value = pending.resolution || resolution.value
      duration.value = pending.duration || duration.value
      // Re-display the uploaded reference image(s) after a reload. They're
      // served (cookie-authed) from /images; re-fetch into data URLs so the
      // thumbnails show AND the refs stay re-submittable if the user regenerates.
      restoreRefs(pending.reference_urls)
    }
    return
  }
  // No pending. If our locally-shown job just finished on the server (same id,
  // status flipped to success/failed), promote it to the result view — this is
  // the live "I'm watching my own generation finish" case and stays.
  if (current.value && current.value.status === 'pending' && latest && latest.id === current.value.id) {
    current.value = { ...latest, status: latest.status === 'success' ? 'done' : latest.status }
    if (latest.url) current.value.url = latest.url
    busy.value = false
    statusText.value = ''
    refreshMe()
    return
  }
  // Intentionally NO restore of an already-finished result on first paint /
  // navigation: the playground only ever shows an in-progress job (or the one
  // that just completed while watched). Past results live in /记录 (logs), not
  // re-echoed onto a freshly opened workspace.
}

function onKey(e) { if (e.key === 'Escape') lightbox.value = null }

onMounted(async () => {
  refreshMe()   // pull the latest real balance
  const [mm, pp] = await Promise.all([api('/managed-models'), api('/video-presets')])
  allModels.value = mm.data?.data || []
  presets.value = pp.data?.data || []
  // Pre-fill from query string (?prompt=...&model=...) — used by the home
  // page's example cards to seed the form in one click.
  const qPrompt = String(route.query.prompt || '')
  const qModel = String(route.query.model || '')
  if (qPrompt) prompt.value = qPrompt
  let selected = null
  if (qModel) {
    selected = allModels.value.find((m) => m.id === qModel && m.enabled !== false)
  }
  // If the draft already points at a still-available model AND no fresher
  // intent came in from the URL, keep the draft as-is. Otherwise fall back to
  // the first usable model.
  const draftModel = !qModel && modelId.value
    ? allModels.value.find((m) => m.id === modelId.value && m.enabled !== false)
    : null
  if (!selected && !draftModel) {
    selected = allModels.value.find((m) => m.enabled !== false && m.type === 'image')
      || allModels.value.find((m) => m.enabled !== false)
  }
  // Always re-apply defaults — even when restoring the persisted draft model —
  // so a stale ratio/resolution that the model no longer supports (e.g. a saved
  // "2K" for a model that's now 1K-only) is normalized to a valid, priced tier
  // instead of being sent as-is and rejected with "unsupported or unpriced".
  const chosen = selected || draftModel
  if (chosen) {
    mode.value = chosen.type
    modelId.value = chosen.id
    applyModelDefaults()
  }
  window.addEventListener('keydown', onKey)
  // Restore any in-flight or recently-finished job for this user, then poll
  // every 2s so a parallel tab / device sees changes within one tick.
  poll()
  pollTimer = setInterval(poll, 2000)
})
onUnmounted(() => {
  window.removeEventListener('keydown', onKey)
  clearInterval(pollTimer)
})
</script>

<template>
  <section class="theme-text grid lg:grid-cols-[420px_1fr] gap-6">
    <!-- LEFT: controls — every interactive element accepts :disabled="busy"
         so the form locks the moment a generation kicks off. Reload, parallel
         tab and tab-switch all see the same locked state via poll(). -->
    <div class="card p-5 space-y-5 lg:sticky lg:top-24 self-start">
      <!-- mode switch -->
      <div class="grid grid-cols-2 gap-2 p-1 bg-slate-100 rounded-xl">
        <button @click="setMode('image')" type="button" :disabled="busy"
                class="rounded-lg py-2 text-sm font-medium transition-colors disabled:cursor-not-allowed"
                :class="mode === 'image' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500 hover:text-slate-700'">
          <Icon name="files" class="w-4 h-4 inline -mt-0.5" /> 生图
        </button>
        <button @click="setMode('video')" type="button" :disabled="busy"
                class="rounded-lg py-2 text-sm font-medium transition-colors disabled:cursor-not-allowed"
                :class="mode === 'video' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500 hover:text-slate-700'">
          <Icon name="video" class="w-4 h-4 inline -mt-0.5" /> 生视频
        </button>
      </div>

      <!-- model -->
      <div>
        <label class="block text-xs font-medium text-slate-500 mb-1.5">模型</label>
        <SelectMenu v-if="models.length" :model-value="modelId" @update:model-value="selectModel"
                    :options="modelOptions" placeholder="选择模型" mono :disabled="busy" />
        <div v-else class="rounded-lg border border-dashed border-slate-200 px-3 py-4 text-xs text-slate-400 text-center">
          还没有可用的{{ mode === 'video' ? '视频' : '图像' }}模型 ·
          <router-link to="/admin/models" class="text-slate-700 underline">去添加</router-link>
        </div>
      </div>

      <!-- prompt -->
      <div>
        <label class="block text-xs font-medium text-slate-500 mb-1.5">提示词</label>
        <textarea v-model="prompt" rows="4" :disabled="busy" class="field resize-none disabled:opacity-60 disabled:cursor-not-allowed"
                  placeholder="描述想要的画面…如：黄昏时分,金色麦田里奔跑的金毛猎犬,电影感"></textarea>
      </div>

      <!-- ratio + res + duration. Single-option controls are hidden — the
           value is still set from the model's defaults and sent to the API,
           so the user doesn't have to acknowledge a choice they don't have. -->
      <div v-if="ratios.length > 0 && showRatio">
        <label class="block text-xs font-medium text-slate-500 mb-1.5">比例</label>
        <div class="flex flex-wrap gap-1.5">
          <button v-for="r in ratios" :key="r" type="button" @click="ratio = r" :disabled="busy"
                  class="rounded-lg px-3 py-1.5 text-xs font-medium transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                  :class="ratio === r ? 'bg-slate-900 text-white' : 'bg-slate-100 text-slate-600 hover:bg-slate-200'">
            {{ r }}
          </button>
        </div>
      </div>

      <div v-if="resolutions.length > 0">
        <label class="block text-xs font-medium text-slate-500 mb-1.5">{{ mode === 'video' ? '分辨率' : '画质' }}</label>
        <div class="flex flex-wrap gap-1.5">
          <button v-for="r in resolutions" :key="r" type="button" @click="resolution = r" :disabled="busy"
                  class="rounded-lg px-3 py-1.5 text-xs font-medium transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                  :class="resolution === r ? 'bg-slate-900 text-white' : 'bg-slate-100 text-slate-600 hover:bg-slate-200'">
            {{ r }}
          </button>
        </div>
      </div>

      <div v-if="mode === 'video' && durations.length > 0">
        <label class="block text-xs font-medium text-slate-500 mb-1.5">时长</label>
        <div class="flex flex-wrap gap-1.5">
          <button v-for="d in durations" :key="d" type="button" @click="duration = d" :disabled="busy"
                  class="rounded-lg px-3 py-1.5 text-xs font-medium transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                  :class="duration === d ? 'bg-slate-900 text-white' : 'bg-slate-100 text-slate-600 hover:bg-slate-200'">
            {{ d }}
          </button>
        </div>
      </div>

      <!-- reference images -->
      <div v-if="maxRefs > 0">
        <label class="block text-xs font-medium text-slate-500 mb-1.5">
          参考图
          <span class="text-slate-400 font-normal">
            (最多 {{ maxRefs }} 张{{ refMode === 'frame' && mode === 'video' ? (maxRefs >= 2 ? ' · 首帧/末帧' : ' · 首帧') : '' }} · 单张 ≤8MB)
          </span>
          <span v-if="refsRequired" class="text-rose-500">*</span>
        </label>
        <div class="flex gap-2 flex-wrap items-start">
          <div v-for="(img, i) in refImages" :key="i"
               class="relative w-20 h-20 rounded-lg overflow-hidden border border-slate-200 bg-slate-50 transition-all"
               :class="busy ? 'opacity-60 grayscale pointer-events-none' : ''">
            <img :src="img.dataUrl || img.url" class="w-full h-full object-cover" />
            <button type="button" @click="removeRef(i)" :disabled="busy"
                    class="absolute top-1 right-1 w-5 h-5 rounded-full bg-slate-900/70 text-white hover:bg-rose-500 grid place-items-center disabled:opacity-40 disabled:cursor-not-allowed">
              <Icon name="close" class="w-3 h-3" />
            </button>
            <div v-if="refMode === 'frame' && mode === 'video' && maxRefs >= 2"
                 class="absolute bottom-0 inset-x-0 text-[10px] text-white bg-slate-900/60 text-center py-0.5">
              {{ i === 0 ? '首帧' : (i === 1 ? '末帧' : '') }}
            </div>
          </div>
          <button v-if="refImages.length < maxRefs" type="button" @click="openPicker" :disabled="busy"
                  class="w-20 h-20 rounded-lg border-2 border-dashed border-slate-200 text-slate-400 hover:bg-slate-50 hover:border-slate-300 grid place-items-center disabled:opacity-40 disabled:cursor-not-allowed">
            <Icon name="plus" class="w-5 h-5" />
          </button>
        </div>
        <input ref="fileInput" type="file" accept="image/*" multiple class="hidden" @change="onFiles" />
      </div>

      <button @click="run" :disabled="busy || !models.length || price == null || !canAfford"
              class="btn-primary w-full !py-3 flex items-center justify-center gap-2 leading-none">
        <Icon name="spark" class="w-4 h-4 shrink-0" />
        <span class="leading-none">{{ busy ? (mode === 'video' ? '生成中…请耐心等待' : '生成中…') : '生成' }}</span>
        <span v-if="!busy && price != null" class="text-xs opacity-70 tabular-nums leading-none">· {{ priceLabel }}</span>
        <span v-if="!busy && price != null && !canAfford" class="text-xs text-rose-200 leading-none">积分不足</span>
      </button>

      <!-- Validation / upload errors (model/prompt/ref/price/credits/oversized
           image). The `error` ref had no render target before, so these messages
           were silently swallowed. -->
      <p v-if="error" class="text-xs text-rose-500 break-all">{{ error }}</p>

    </div>

    <!-- RIGHT: single latest result (replaces on each new generation).
         min-w-0: the 1fr grid track defaults to min-width:auto, so a long
         unbroken prompt would otherwise blow the column wider than the page
         (truncate can't shrink a track that won't shrink). -->
    <div class="space-y-4 min-w-0">
      <div v-if="!current && !busy"
           class="card p-14 grid place-items-center text-slate-400 text-center">
        <span class="w-16 h-16 rounded-2xl bg-slate-100 grid place-items-center mb-4">
          <Icon name="spark" class="w-7 h-7 text-slate-400" />
        </span>
        <p class="text-sm">还没有生成过 — 在左侧写提示词,点击「生成」</p>
        <router-link to="/logs" class="text-xs text-slate-500 hover:text-white mt-3 transition-colors">查看历史记录 →</router-link>
      </div>

      <div v-else-if="current" class="card overflow-hidden">
        <div class="px-5 py-3 border-b border-slate-100 flex items-center justify-between gap-3">
          <div class="min-w-0">
            <div class="text-sm font-medium line-clamp-2 break-words">{{ current.prompt }}</div>
            <div class="text-[11px] text-slate-400 mt-0.5 font-mono">
              {{ current.model }} · {{ current.ratio }} · {{ current.resolution }}
              <span v-if="current.kind === 'video'"> · {{ current.duration }}</span>
              <span v-if="current.elapsed_ms"> · {{ (current.elapsed_ms / 1000).toFixed(1) }}s</span>
            </div>
          </div>
          <!-- only when a finished result exists — hidden while pending/failed -->
          <div v-if="current.url && current.status !== 'pending' && current.status !== 'failed'"
               class="flex items-center gap-1.5 shrink-0">
            <a :href="current.url" :download="''" class="btn-soft" title="下载">
              <Icon name="download" class="w-3.5 h-3.5" />
            </a>
            <button @click="copyLink(current.url)" class="btn-soft" title="复制链接">
              <Icon name="copy" class="w-3.5 h-3.5" />
            </button>
          </div>
        </div>

        <div class="bg-slate-50 grid place-items-center min-h-[260px]">
          <div v-if="current.status === 'pending'" class="text-sm text-slate-400 py-12 flex flex-col items-center gap-2">
            <span class="w-10 h-10 rounded-xl bg-white grid place-items-center animate-pulse">
              <Icon name="spark" class="w-4 h-4 text-slate-400" />
            </span>
            {{ statusText || '生成中…' }}
          </div>
          <div v-else-if="current.status === 'failed'" class="text-sm text-rose-600 py-12 px-5 max-w-xl text-center">
            <div class="font-medium mb-1">生成失败</div>
            <div class="text-xs text-rose-500 break-all">{{ current.error }}</div>
          </div>
          <template v-else>
            <video v-if="current.kind === 'video'" :src="current.url" controls
                   class="max-w-full max-h-[600px] object-contain" />
            <img v-else :src="current.url" @click="lightbox = current"
                 class="max-w-full max-h-[600px] object-contain cursor-zoom-in" />
          </template>
        </div>
      </div>
    </div>

    <!-- Lightbox — shared component, consistent with 图片管理 / 日志 -->
    <MediaLightbox
      v-if="lightbox"
      :src="lightbox.url"
      :kind="lightbox.kind"
      :prompt="lightbox.prompt"
      :meta="[lightbox.model, lightbox.ratio, lightbox.resolution, (lightbox.kind === 'video' ? lightbox.duration : '')].filter(Boolean).join(' · ')"
      :download-name="(lightbox.url || '').split('/').pop()"
      @close="lightbox = null" />

    <!-- Toast -->
    <transition name="fade">
      <div v-if="toast"
           class="fixed bottom-6 left-1/2 -translate-x-1/2 z-[60] bg-slate-900 text-white text-xs px-4 py-2 rounded-lg shadow-lg">
        {{ toast }}
      </div>
    </transition>
  </section>
</template>

<style scoped>
.fade-enter-active, .fade-leave-active { transition: opacity 0.15s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
