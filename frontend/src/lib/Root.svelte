<script>
    import { onDestroy, onMount } from "svelte";
    import Node from "./Node.svelte";

    let treeData = $state([]);
    let hasData = $derived(treeData.length > 0);

    let intervalId = $state(0);
    let currentRequestId = $state(0);

    // Fetch data from the API
    onMount(async () => {
        intervalId = setInterval(fetchRequests, 1000);
    });

    onDestroy(async () => {
        clearInterval(intervalId);
    });

    function snakeToTitleCase(str) {
        return str
            .split("_")
            .map(
                (word) =>
                    word.charAt(0).toUpperCase() + word.slice(1).toLowerCase(),
            )
            .join(" ");
    }

    async function fetchRequests() {
        const response = await fetch("http://192.168.106.19:2020/request");
        if (!response.ok) {
            return;
        }

        const data = await response.json();
        if (!data.has_request) {
            treeData = [];
            return;
        }
        if (currentRequestId == data.id) {
            return;
        }

        currentRequestId = data.id;
        treeData = Object.entries(data.request);
    }

    async function resumeRequest() {
        const response = await fetch("http://192.168.106.19:2020", {
            method: "POST",
        });
        if (response.ok) {
            fetchRequests();
        }
    }
</script>

{#if hasData}
    <div class="container" id="content">
        <div class="top-row">
            <button id="resumeButton" onclick={resumeRequest}
                >Resume &#9658;
            </button>
        </div>

        <div class="tree json-tree">
            {#each treeData as [key, node]}
                <Node key={snakeToTitleCase(key)} {node} />
            {/each}
        </div>
    </div>
{:else}
    <div class="loading" id="loading">Waiting for request...</div>
{/if}

<style>
    .tree {
        font-family: Arial, sans-serif;
    }
</style>
