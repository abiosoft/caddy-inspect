<script>
    import { onDestroy, onMount } from "svelte";
    import Node from "./Node.svelte";

    const host = "http://192.168.106.19:2020";

    let treeData = $state([]);
    let hasData = $derived(treeData.length > 0);
    let requestUrl = $state("");

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
        const response = await fetch(`${host}/request`);
        if (!response.ok) {
            return;
        }

        const data = await response.json();
        if (!data.has_request) {
            treeData = [];
            document.title = "Caddy Inspect";
            return;
        }
        if (currentRequestId == data.id) {
            return;
        }

        currentRequestId = data.id;
        treeData = Object.entries(data.request);
        requestUrl = data.request.url;

        window.focus();
        document.title = "Caddy Inspect - " + requestUrl;
    }

    async function resumeRequest() {
        const response = await fetch(host, {
            method: "POST",
        });
        if (response.ok) {
            fetchRequests();
        }
    }
    async function stopRequest() {
        const response = await fetch(`${host}/stop`, {
            method: "POST",
        });
        if (response.ok) {
            fetchRequests();
        }
    }
</script>

<div class="container" id="content">
    <div class="header">Caddy Inspect</div>
    <hr />
    <br />
    {#if hasData}
        <div class="top-row">
            <button
                id="resumeButton"
                onclick={resumeRequest}
                title="Resume the request"
                >Resume &#9658;
            </button>
            <button
                id="stopButton"
                class="danger"
                onclick={stopRequest}
                title="Terminate the request"
                >Stop &#9632;
            </button>
        </div>

        <div class="tree json-tree">
            {#each treeData as [key, node]}
                <Node key={snakeToTitleCase(key)} {node} />
            {/each}
        </div>
    {:else}
        <div class="loading" id="loading">Waiting for request...</div>
    {/if}
</div>

<style>
    .tree {
        font-family: Arial, sans-serif;
    }
</style>
