<script>
    // @ts-nocheck

    import { onDestroy, onMount } from "svelte";
    import Node from "./Node.svelte";

    const host = "http://127.0.0.1:2020";

    let treeData = $state([]);
    let hasData = $derived(treeData.length > 0);
    let requestUrl = $state("");

    let intervalId = $state(0);
    let currentRequestId = $state(0);
    let hasResponse = $state(false);

    // Fetch data from the API
    onMount(async () => {
        intervalId = setInterval(fetchRequests, 1000);
    });

    onDestroy(async () => {
        clearInterval(intervalId);
    });

    $effect(() => {
        document.title = !!requestUrl
            ? `Caddy Inspect - ${requestUrl}`
            : "Caddy Inspect";
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

    function resetRequests() {
        treeData = [];
        requestUrl = "";
        currentRequestId = 0;
    }

    async function fetchRequests() {
        const response = await fetch(`${host}/request`);
        if (!response.ok) {
            return;
        }

        const data = await response.json();
        if (!data.has_request) {
            resetRequests();
            return;
        }
        if (currentRequestId == data.id) {
            return;
        }

        currentRequestId = data.id;
        treeData = Object.entries(data.request);
        requestUrl = data.request.url;
        hasResponse = data.has_response;

        window.focus();
    }

    async function resumeRequest() {
        const response = await fetch(host, {
            method: "POST",
        });
        if (response.ok) {
            resetRequests();
            fetchRequests();
        }
    }
    async function stopRequest() {
        const response = await fetch(`${host}/stop`, {
            method: "POST",
        });
        if (response.ok) {
            resetRequests();
            fetchRequests();
        }
    }

    async function stepRequest() {
        const response = await fetch(`${host}/step`, {
            method: "POST",
        });
        if (response.ok) {
            resetRequests();
            fetchRequests();
        }
    }
</script>

<div class="container" id="content">
    <div class="header">
        <img class="logo" alt="Caddy logo" /> <span>Inspect</span>
    </div>
    <hr />
    {#if hasData}
        <div class="top-row">
            <span class="top-row-item">
                <button
                    id="resumeButton"
                    onclick={resumeRequest}
                    title="Resume the request"
                    >&#x23f5;
                </button>
            </span>
            {#if !hasResponse}
                <span class="top-row-item">
                    <button
                        id="resumeButton"
                        onclick={stepRequest}
                        title="Resume the request but pause to inspect the response"
                        >&#x23ed;
                    </button>
                </span>
            {/if}
            <span class="top-row-item top-row-right">
                <button
                    id="stopButton"
                    class="danger"
                    onclick={stopRequest}
                    title="Terminate the request"
                    >&#x23f9;
                </button>
            </span>
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
