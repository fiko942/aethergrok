<script lang="ts">
  let { active = false }: { active?: boolean } = $props();
</script>

{#if active}
  <div
    class="pointer-events-none fixed inset-0 z-[9999] animate-snapshot-flash bg-white"
    aria-hidden="true"
  ></div>
{/if}

<style>
  @keyframes flashOut {
    0% {
      opacity: 0.85;
    }
    50% {
      opacity: 0.45;
    }
    100% {
      opacity: 0;
    }
  }

  .animate-snapshot-flash {
    animation: flashOut 240ms cubic-bezier(0.1, 0.9, 0.2, 1) forwards;
  }
</style>
