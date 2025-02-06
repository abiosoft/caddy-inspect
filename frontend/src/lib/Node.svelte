<script>
    import { onMount } from "svelte";
    import Node from "./Node.svelte";

    const ARRAY_KEY = "__array__";

    let { key, node } = $props();

    let nodeEntries = Object.entries(node);
    let nodeIsArray = $derived(Array.isArray(node));
    let hasChildren = $derived.by(() => {
        console.log("derived by", key, typeof node, node);
        return typeof node === "object" && node !== null;
    });

    let itemIsArray = key == ARRAY_KEY;

    let isOpen = $state(false);

    function toggle() {
        isOpen = !isOpen;
    }

    onMount(() => {
        $inspect(node);
    });
</script>

<div class="tree-node">
    <div class="node-header" onclick={toggle}>
        {#if hasChildren}
            <span class="toggle-icon">{isOpen ? "▼" : "▶"}</span>
        {:else}
            <span class="toggle-icon">&nbsp; &nbsp;</span>
        {/if}
        <div class="key-value-container">
            {#if itemIsArray}
                <input class="value-only" value={node} readonly="readonly" />
            {:else}
                <span class="key">{key}</span>
                {#if !hasChildren}
                    <input class="value" value={node} readonly="readonly" />
                {/if}
            {/if}
        </div>
    </div>

    {#if hasChildren && isOpen}
        <div class="children">
            {#if nodeIsArray}
                {#each node as node}
                    <Node key={ARRAY_KEY} {node} />
                {/each}
            {:else}
                {#each nodeEntries as [key, node]}
                    <Node {key} {node} />
                {/each}
            {/if}
        </div>
    {/if}
</div>

<style>
    .tree-node {
        margin-left: 20px;
        display: flex;
        flex-direction: column;
    }
    .node-header {
        cursor: pointer;
        display: flex;
        align-items: center;
    }
    .toggle-icon {
        margin-right: 5px;
        font-size: 10px;
    }

    .key-value-container {
        display: flex;
        justify-content: space-between;
        width: 100%;
    }

    .key {
        flex: 1;
        margin-right: 10px;
        font-family: monospace;
        padding: 5px;
        font-weight: bold;
    }

    .value,
    .value-only {
        flex: 2;
        margin-left: 10px;
        border: none;
        background: transparent;
        font-family: monospace;
        font-size: inherit;
        color: inherit;
        padding: 5px;
        border-radius: 4px;
    }

    /* Make .value-only take full width if it is the only child */
    .key-value-container:has(.value-only):not(:has(.key)) .value-only {
        flex: 1 1 100%;
        margin-left: 0; /* Remove left margin when alone */
    }

    .children {
        margin-left: 20px;
    }
</style>
