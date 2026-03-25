<script lang="ts">
  import { onMount } from "svelte";
  import { toast } from "svelte-sonner";

  import { EventsOn } from "@wailsjs/runtime/runtime";
  import { MoveToTrash } from "@wailsjs/go/main/App";
  import InputArea from "./InputArea.svelte";
  import Controls from "./Controls.svelte";
  import ResultsPane from "./ResultsPane.svelte";

  interface DedupGroup {
    Size: number;
    Hash: string;
    Paths: string[];
  }

  let is_scanning: boolean = $state(false);
  let is_cleaning: boolean = $state(false);
  let input_files: string[] = $state([]);
  let results: DedupGroup[] = $state([]);
  let keep_selections: Record<string, string> = $state({});

  function getKeep(group: DedupGroup): string {
    return keep_selections[group.Hash] ?? group.Paths[0];
  }

  async function cleanAll() {
    const toTrashByGroup: { hash: string; keep: string; paths: string[] }[] =
      results.map((g) => {
        const keep = getKeep(g);
        return { hash: g.Hash, keep, paths: g.Paths.filter((p) => p !== keep) };
      });

    const allToTrash = toTrashByGroup.flatMap((g) => g.paths);
    if (allToTrash.length === 0) return;

    is_cleaning = true;
    try {
      const trashResults = await MoveToTrash(allToTrash);
      const failedPaths = new Set(
        trashResults.filter((r) => r.error).map((r) => r.path),
      );

      if (failedPaths.size === 0) {
        results = [];
        keep_selections = {};
        input_files = [];
        toast.success("All duplicates moved to Trash.");
      } else {
        const newResults: DedupGroup[] = [];
        for (const g of toTrashByGroup) {
          const original = results.find((r) => r.Hash === g.hash)!;
          const remaining = [
            g.keep,
            ...g.paths.filter((p) => failedPaths.has(p)),
          ];
          if (remaining.length >= 2) {
            newResults.push({ ...original, Paths: remaining });
          }
        }
        results = newResults;
        toast.warning(
          `${failedPaths.size} file${failedPaths.size > 1 ? "s" : ""} could not be moved to Trash.`,
        );
      }
    } catch (e) {
      toast.error(`Failed to move files to Trash: ${e}`);
    } finally {
      is_cleaning = false;
    }
  }

  onMount(() => {
    const offComplete = EventsOn(
      "dedup:complete",
      (groups: DedupGroup[] | null) => {
        is_scanning = false;
        keep_selections = {};
        results = groups ?? [];
        if (results.length === 0) {
          toast.info("No duplicates found.");
        }
      },
    );
    const offError = EventsOn("dedup:error", (payload: { error: string }) => {
      is_scanning = false;
      toast.error(`Scan failed: ${payload?.error ?? "unknown error"}`);
    });
    const offCancelled = EventsOn("dedup:cancelled", () => {
      is_scanning = false;
    });

    return () => {
      offComplete();
      offError();
      offCancelled();
    };
  });
</script>

<div class="flex h-full w-full rounded-xl gap-1 px-6 items-center justify-center py-4">
  <InputArea bind:input_files />
  <Controls bind:input_files bind:is_scanning bind:is_cleaning bind:results {cleanAll} />
  <ResultsPane bind:results bind:is_cleaning bind:keep_selections />
</div>
