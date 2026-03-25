<script lang="ts">
  import { onMount } from "svelte";
  import { EventsOn } from "@wailsjs/runtime/runtime";
  import { DisplayFileDialog } from "@wailsjs/go/main/App";

  let { input_files = $bindable([]) }: { input_files: string[] } = $props();

  let is_drag_over: boolean = $state(false);
  let drag_counter: number = 0;

  onMount(() => {
    EventsOn(
      "wails:file-drop",
      async (x: number, y: number, paths: string[]) => {
        // Check if the drop occurred within our input area
        if (is_drag_over) {
          input_files = paths;
        }
        is_drag_over = false;
        drag_counter = 0;
      },
    );
  });

  async function onInputAreaClicked() {
    var file_paths: string[] = [];
    try {
      file_paths = await DisplayFileDialog();
    } catch (error: any) {
      // Toast.error(
      //     "An error occurred while opening the file selection dialog.",
      // );
      return;
    }
    input_files = file_paths;
  }

  function handleDragEnter(event: DragEvent) {
    event.preventDefault();
    drag_counter++;
    if (drag_counter === 1) {
      is_drag_over = true;
    }
  }

  function handleDragOver(event: DragEvent) {
    event.preventDefault();
  }

  function handleDragLeave(event: DragEvent) {
    event.preventDefault();
    drag_counter--;
    if (drag_counter === 0) {
      is_drag_over = false;
    }
  }

  function handleDrop(event: DragEvent) {
    event.preventDefault();
    // Note: We don't handle the drop here, the wails event listener will handle it to obtain the full file paths.
  }
</script>

<div class={`w-64 h-full overflow-hidden rounded-xl flex flex-col ${input_files?.length > 0 ? "bg-card" : ""}`}>
  {#if input_files?.length > 0}
    <h2 class="text-lg font-semibold shrink-0 text-center pt-4 pb-2">
      Selected Files ({input_files.length})
    </h2>
    <ul class="overflow-y-auto flex-1 text-muted-foreground text-sm space-y-3 px-3 pb-2">
      {#each input_files as sourcePath}
        <li
          class="truncate whitespace-nowrap overflow-hidden rounded"
          title={sourcePath}
        >
          {sourcePath.split("/").pop()}
        </li>
      {/each}
    </ul>
  {:else}
    <div class="h-full w-full flex flex-col p-4">
      <div
        class="h-full border-2 border-dashed border-border rounded-lg transition-colors duration-200 flex flex-col items-center justify-center p-8 cursor-pointer relative"
        class:bg-accent={is_drag_over}
        ondragenter={handleDragEnter}
        ondragover={handleDragOver}
        ondragleave={handleDragLeave}
        ondrop={handleDrop}
        onclick={input_files.length === 0 ? onInputAreaClicked : undefined}
      >
        <div class="text-center">
          <div
            class="w-16 h-16 mx-auto mb-4 rounded-full flex items-center justify-center flex-shrink-0"
          >
            <svg
              class="w-8 h-8"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"
              />
            </svg>
          </div>
          <h3 class="text-lg font-semibold mb-2">Drop your files here</h3>
          <p class="mb-4">or click to browse</p>
        </div>
      </div>
    </div>
  {/if}
</div>
