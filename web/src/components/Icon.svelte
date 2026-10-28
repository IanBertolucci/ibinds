<script lang="ts">
  let { name, size = 18 }: { name: string; size?: number } = $props();
</script>

<svg
  width={size}
  height={size}
  viewBox="0 0 24 24"
  fill="none"
  stroke="currentColor"
  stroke-width="1.8"
  stroke-linecap="round"
  stroke-linejoin="round"
  aria-hidden="true"
>
  {#if name === 'search'}
    <circle cx="11" cy="11" r="7" />
    <path d="m21 21-4.35-4.35" />
  {:else if name === 'settings'}
    <path d="M4 7h16" />
    <path d="M4 17h16" />
    <circle cx="9" cy="7" r="2" />
    <circle cx="15" cy="17" r="2" />
  {:else if name === 'edit'}
    <path d="M12 20h9" />
    <path d="M16.5 3.5a2.12 2.12 0 0 1 3 3L7 19l-4 1 1-4Z" />
  {:else if name === 'copy'}
    <rect x="9" y="9" width="11" height="11" rx="2" />
    <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
  {:else if name === 'close'}
    <path d="M18 6 6 18" />
    <path d="m6 6 12 12" />
  {:else if name === 'check'}
    <path d="m20 6-11 11-5-5" />
  {:else if name === 'chevron'}
    <path d="m6 9 6 6 6-6" />
  {:else if name === 'device'}
    <rect x="3" y="7" width="18" height="10" rx="3" />
    <path d="M7 12h2" />
    <path d="M17 10v4" />
    <path d="M15 12h4" />
  {:else if name === 'refresh'}
    <path d="M21 12a9 9 0 1 1-2.64-6.36" />
    <path d="M21 3v6h-6" />
  {/if}
</svg>
