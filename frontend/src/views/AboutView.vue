<script setup>
import { computed } from 'vue'
import { site } from '../site'

const contact = computed(() => site.contact || {})
</script>

<template>
  <div class="space-y-12 max-w-3xl">
    <!-- intro -->
    <section>
      <div class="text-[10px] uppercase tracking-[0.3em] text-fuchsia-300/70 font-medium">关于</div>
      <h1 class="mt-2 text-4xl md:text-5xl font-bold tracking-tight text-[color:var(--fg)]">关于 {{ site.title }}</h1>
      <p class="text-[color:var(--fg-2)] mt-5 text-base md:text-lg leading-relaxed max-w-xl">
        {{ site.title }} 是一个聚合 GPT、Gemini、Firefly、Flux 等顶级模型的 AI 生图与生视频平台。
        把脑海里的画面写成一句话,交给我们替你变成图像与视频。
      </p>
    </section>

    <!-- contact -->
    <section>
      <h2 class="text-2xl font-bold tracking-tight text-[color:var(--fg)] mb-1">联系我们</h2>
      <p class="text-[color:var(--fg-3)] mb-6">有问题、合作,或想加入交流群?随时找我们。</p>

      <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
        <a v-if="contact.shop" :href="contact.shop" target="_blank" rel="noopener"
           class="group flex items-center justify-between gap-3 rounded-2xl bg-[var(--surface)] ring-1 ring-[color:var(--hairline)] p-5 hover:ring-[color:var(--fg-faint)] transition-all">
          <div>
            <div class="text-[10px] uppercase tracking-[0.25em] text-fuchsia-300/80">商店</div>
            <div class="text-base font-semibold text-[color:var(--fg)] mt-1 group-hover:text-fuchsia-400 transition-colors">前往充值商店</div>
          </div>
          <span class="text-[color:var(--fg-faint)] group-hover:translate-x-1 transition-transform">→</span>
        </a>

        <a v-if="contact.email" :href="`mailto:${contact.email}`"
           class="group rounded-2xl bg-[var(--surface)] ring-1 ring-[color:var(--hairline)] p-5 hover:ring-[color:var(--fg-faint)] transition-all">
          <div class="text-[10px] uppercase tracking-[0.25em] text-[color:var(--fg-3)]">邮箱</div>
          <div class="text-base font-semibold text-[color:var(--fg)] mt-1 break-all group-hover:text-sky-400 transition-colors">{{ contact.email }}</div>
        </a>

        <a v-if="contact.qq && contact.qq_link" :href="contact.qq_link" target="_blank" rel="noopener"
           class="group flex items-center justify-between gap-3 rounded-2xl bg-[var(--surface)] ring-1 ring-[color:var(--hairline)] p-5 hover:ring-[color:var(--fg-faint)] transition-all">
          <div>
            <div class="text-[10px] uppercase tracking-[0.25em] text-[color:var(--fg-3)]">QQ</div>
            <div class="text-base font-semibold text-[color:var(--fg)] mt-1 tabular-nums break-all group-hover:text-sky-400 transition-colors">{{ contact.qq }}</div>
          </div>
          <span class="text-[color:var(--fg-faint)] group-hover:translate-x-1 transition-transform">→</span>
        </a>
        <div v-else-if="contact.qq" class="rounded-2xl bg-[var(--surface)] ring-1 ring-[color:var(--hairline)] p-5">
          <div class="text-[10px] uppercase tracking-[0.25em] text-[color:var(--fg-3)]">QQ</div>
          <div class="text-base font-semibold text-[color:var(--fg)] mt-1 tabular-nums break-all">{{ contact.qq }}</div>
        </div>

        <a v-if="contact.qq_group && contact.qq_group_link" :href="contact.qq_group_link" target="_blank" rel="noopener"
           class="group flex items-center justify-between gap-3 rounded-2xl bg-[var(--surface)] ring-1 ring-[color:var(--hairline)] p-5 hover:ring-[color:var(--fg-faint)] transition-all">
          <div>
            <div class="text-[10px] uppercase tracking-[0.25em] text-[color:var(--fg-3)]">QQ 群</div>
            <div class="text-base font-semibold text-[color:var(--fg)] mt-1 tabular-nums break-all group-hover:text-emerald-400 transition-colors">{{ contact.qq_group }}</div>
          </div>
          <span class="text-[color:var(--fg-faint)] group-hover:translate-x-1 transition-transform">→</span>
        </a>
        <div v-else-if="contact.qq_group" class="rounded-2xl bg-[var(--surface)] ring-1 ring-[color:var(--hairline)] p-5">
          <div class="text-[10px] uppercase tracking-[0.25em] text-[color:var(--fg-3)]">QQ 群</div>
          <div class="text-base font-semibold text-[color:var(--fg)] mt-1 tabular-nums break-all">{{ contact.qq_group }}</div>
        </div>
      </div>

      <p v-if="!contact.shop && !contact.email && !contact.qq && !contact.qq_group"
         class="text-sm text-[color:var(--fg-3)]">管理员尚未配置联系方式。</p>
    </section>
  </div>
</template>
