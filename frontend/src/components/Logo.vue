<script setup>
// Vivid mark: a violet→fuchsia rounded badge with a faceted, folded "V".
// Self-contained SVG so it renders identically on dark and light shells.
// Each instance gets unique gradient ids (useId) so multiple logos on one
// page don't share/clobber defs.
import { useId } from 'vue'

defineProps({ size: { type: [Number, String], default: 40 } })

const uid = useId()
const gBg = `vbg-${uid}`
const gGloss = `vgl-${uid}`
</script>

<template>
  <svg :width="size" :height="size" viewBox="0 0 48 48" fill="none"
       xmlns="http://www.w3.org/2000/svg" class="shrink-0">
    <defs>
      <linearGradient :id="gBg" x1="6" y1="4" x2="42" y2="44" gradientUnits="userSpaceOnUse">
        <stop stop-color="#a855f7" />
        <stop offset="0.55" stop-color="#7c3aed" />
        <stop offset="1" stop-color="#ec4899" />
      </linearGradient>
      <linearGradient :id="gGloss" x1="24" y1="2" x2="24" y2="30" gradientUnits="userSpaceOnUse">
        <stop stop-color="#ffffff" stop-opacity="0.28" />
        <stop offset="1" stop-color="#ffffff" stop-opacity="0" />
      </linearGradient>
    </defs>

    <!-- badge -->
    <rect x="2" y="2" width="44" height="44" rx="13" :fill="`url(#${gBg})`" />
    <!-- top gloss for depth -->
    <rect x="2" y="2" width="44" height="22" rx="13" :fill="`url(#${gGloss})`" />

    <!-- faceted V: left arm brighter, right arm dimmer -> folded look -->
    <path d="M12 15 H19 L24 26 L24 37 Z" fill="#ffffff" fill-opacity="0.95" />
    <path d="M36 15 H29 L24 26 L24 37 Z" fill="#ffffff" fill-opacity="0.72" />
  </svg>
</template>
