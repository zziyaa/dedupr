<script lang="ts">
  import { Check, Trash2 } from "@lucide/svelte";

  interface DedupGroup {
    Size: number;
    Hash: string;
    Paths: string[];
  }

  let {
    results = $bindable([]),
    is_cleaning = $bindable(false),
    keep_selections = $bindable({}),
  }: {
    results: DedupGroup[];
    is_cleaning: boolean;
    keep_selections: Record<string, string>;
  } = $props();

  function getKeep(group: DedupGroup): string {
    return keep_selections[group.Hash] ?? group.Paths[0];
  }

  function setKeep(group: DedupGroup, path: string) {
    keep_selections[group.Hash] = path;
  }
</script>

<div class={`w-64 h-full overflow-hidden rounded-xl flex flex-col ${results.length > 0 ? "bg-card" : ""}`}>
  {#if results.length > 0}
    <h2 class="text-lg font-semibold shrink-0 text-center pt-4 pb-2">Duplicates</h2>
    <div class="flex flex-col gap-3 overflow-y-auto px-3 pb-2">
      {#each results as group (group.Hash)}
        <div class="rounded-lg border border-border p-2 text-sm">
          <div class="flex items-center justify-between mb-2">
            <span class="text-muted-foreground">
              {group.Paths.length} files · {(group.Size / 1024).toFixed(1)} KB
              each
            </span>
          </div>
          <ul class="space-y-1">
            {#each group.Paths as path}
              {@const kept = path === getKeep(group)}
              <li>
                <button
                  type="button"
                  class="w-full flex items-center gap-2 px-2 py-1 rounded cursor-pointer truncate text-left
                         {kept
                    ? 'border-green-500 bg-green-500/10'
                    : 'text-muted-foreground hover:bg-accent'}"
                  title={path}
                  onclick={() => setKeep(group, path)}
                >
                  {#if kept}
                    <Check class="size-3 shrink-0 text-green-500" />
                  {:else}
                    <Trash2 class="size-3 shrink-0" />
                  {/if}
                  <span class="truncate">{path.split("/").pop()}</span>
                </button>
              </li>
            {/each}
          </ul>
        </div>
      {/each}
    </div>
  {/if}
</div>
