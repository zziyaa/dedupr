<script lang="ts">
  import "./app.css";

  import MainPage from "$lib/components/MainPage.svelte";
  import { Toaster } from "$lib/components/ui/sonner/index.js";

  // preventDefaultContextMenu disables the default browser context menu across the app.
  function preventDefaultContextMenu(event: MouseEvent) {
    event.preventDefault();
    return false;
  }
</script>

<svelte:window on:contextmenu={preventDefaultContextMenu} />

<main
  class="flex flex-col h-screen w-screen select-none cursor-default overflow-hidden"
>
  <!-- Header bar -->
  <div class="w-full h-8 shrink-0" style="--wails-draggable:drag"></div>

  <Toaster />

  <div class="flex-1 min-h-0">
    <MainPage />
  </div>
</main>
