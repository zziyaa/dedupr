<script lang="ts">
  import Button from "$lib/components/ui/button/button.svelte";
  import * as AlertDialog from "$lib/components/ui/alert-dialog/index.js";
  import { RotateCcw, Radar, Trash2 } from "@lucide/svelte";
  import { FindDuplicates, CancelFindDuplicates } from "@wailsjs/go/main/App";

  let {
    input_files = $bindable([]),
    is_scanning = $bindable(false),
    is_cleaning = $bindable(false),
    results = $bindable([]),
    cleanAll = async () => {},
  }: {
    input_files: string[];
    is_scanning: boolean;
    is_cleaning: boolean;
    results: unknown[];
    cleanAll: () => Promise<void>;
  } = $props();

  async function startScan() {
    is_scanning = true;
    try {
      await FindDuplicates(input_files);
    } catch {
      is_scanning = false;
    }
  }

  async function cancelScan() {
    try {
      await CancelFindDuplicates();
    } catch {
      is_scanning = false;
    }
  }

  function clearSelection() {
    input_files = [];
    results = [];
  }
</script>

<div
  class="w-64 h-full flex flex-col rounded-xl items-center justify-center gap-2"
>
  {#if is_scanning}
    <Button
      variant="destructive"
      class="w-2/3 rounded-full px-6 font-semibold text-base py-6"
      onclick={cancelScan}
    >
      Cancel
    </Button>
  {:else}
    {#if results.length === 0}
      <Button
        class="w-2/3 rounded-full px-6 font-semibold text-base py-6"
        disabled={input_files.length < 2 || is_cleaning}
        onclick={startScan}
        title="Start scanning for duplicate files"
      >
        <Radar />
        Scan
      </Button>
    {/if}
    {#if results.length > 0}
      <AlertDialog.Root>
        <AlertDialog.Trigger>
          {#snippet child({ props })}
            <Button
              variant="destructive"
              class="w-2/3 rounded-full px-6 font-semibold text-base py-6"
              disabled={is_cleaning}
              title="Move all duplicate files to Trash"
              {...props}
            >
              <Trash2 />
              Move to Trash
            </Button>
          {/snippet}
        </AlertDialog.Trigger>
        <AlertDialog.Content>
          <AlertDialog.Header>
            <AlertDialog.Title>Are you absolutely sure?</AlertDialog.Title>
            <AlertDialog.Description>
              This will move all duplicate files to the Trash.
            </AlertDialog.Description>
          </AlertDialog.Header>
          <AlertDialog.Footer>
            <AlertDialog.Cancel class='rounded-full'>Cancel</AlertDialog.Cancel>
            <AlertDialog.Action class='rounded-full' onclick={cleanAll}>Continue</AlertDialog.Action>
          </AlertDialog.Footer>
        </AlertDialog.Content>
      </AlertDialog.Root>
    {/if}
    {#if input_files.length > 0}
      <Button
        variant="outline"
        class="w-2/3 rounded-full px-6 font-semibold text-base py-6"
        onclick={clearSelection}
        title="Clear file selection and scan results"
      >
        <RotateCcw />
        Reset
      </Button>
    {/if}
  {/if}
</div>
