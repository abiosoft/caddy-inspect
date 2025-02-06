<script>
    import { onMount } from "svelte";
    import Node from "./Node.svelte";

    let { key, node } = $props();
    let nodeEntries = Object.entries(node);
    let hasChildren = $derived.by(() => {
        console.log("derived by", key, typeof node, node);
        return typeof node === "object" && node !== null;
    });
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
        <span class="key">{key}</span>
        {#if !hasChildren}
            <input class="value" value={node} readonly="readonly" />
        {/if}
    </div>

    {#if hasChildren && isOpen}
        <div class="children">
            {#each nodeEntries as [key, node]}
                <Node {key} {node} />
            {/each}
        </div>
    {/if}
</div>

<style>
    .tree-node {
        margin-left: 20px;
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

    .key {
        margin-right: 10px;
        font-family: monospace;
    }

    .value {
        flex: 1;
        margin-left: 10px;
        border: none;
        background: transparent;
        font-family: monospace;
        font-size: inherit;
        color: inherit;
    }

    .children {
        margin-left: 20px;
    }
</style>
