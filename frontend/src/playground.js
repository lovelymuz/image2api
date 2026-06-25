// Reactive draft of the 画图 form fields. Lives at module scope so the
// values survive PlaygroundView being unmounted (navigation to 首页/记录 etc.)
// and remounted — without this, switching away and back wiped the prompt
// and selected model. Per-tab only; not persisted to localStorage.
import { reactive } from 'vue'

export const draft = reactive({
  mode: '',           // 'image' | 'video'
  modelId: '',
  prompt: '',
  ratio: '',
  resolution: '',
  duration: '',
})

// Copy fields from a server-side job entry (the `/jobs/mine` payload) into
// the draft so a parallel tab can pick up exactly what's being generated.
export function applyJobToDraft(entry) {
  if (!entry) return
  draft.mode = entry.kind === 'video' ? 'video' : 'image'
  draft.modelId = entry.model || ''
  draft.prompt = entry.prompt || ''
  draft.ratio = entry.ratio || ''
  draft.resolution = entry.resolution || ''
  draft.duration = entry.duration || ''
}
