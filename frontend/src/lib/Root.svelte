<script>
    // @ts-nocheck

    import { onDestroy, onMount } from "svelte";
    import Node from "./Node.svelte";

    const host = "http://127.0.0.1:2020";

    let treeData = $state([]);
    let hasData = $derived(treeData.length > 0);
    let requestUrl = $state("");
    let requestMethod = $state("");
    let sourceFile = $state("");
    let sourceLine = $state(0);
    let sourceLineStart = $state(0);
    let hasSourceFile = $derived(sourceFile != "");
    let sources = $state([]);

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
            ? `Caddy Inspect - ${requestMethod.toUpperCase()} ${requestUrl}`
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
        requestMethod = "";
        currentRequestId = 0;
        sourceFile = "";
        sourceLine = 0;
        sourceLineStart = 0;
        sources = [];
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

        // delete caddyfile details from actual data
        const { caddyfile, ...request } = data.request;

        // request id
        currentRequestId = data.id;

        // other rendering properties
        requestUrl = data.request.url;
        requestMethod = data.request.method;
        hasResponse = data.has_response;
        if (caddyfile) {
            sourceFile = caddyfile.file;
            sourceLine = caddyfile.line;
            sourceLineStart = caddyfile.source_line_start;
            sources = caddyfile.source;
        }

        // tree rendering data
        treeData = Object.entries(request);

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
            {#if hasSourceFile}
                <div class="source-file">
                    <code> {sourceFile}:{sourceLine}</code>
                </div>
                <div class="snippet">
                    {#each sources as line, i}
                        <div
                            class={{
                                "code-line": true,
                                highlight: i + sourceLineStart == sourceLine,
                            }}
                        >
                            <span class="mark">
                                {#if i + sourceLineStart == sourceLine}
                                    {#if hasResponse}
                                        &uarr;
                                    {:else}
                                        &darr;
                                    {/if}
                                    &#9673
                                {/if}
                            </span>
                            <span class="line-number"
                                >{i + sourceLineStart}</span
                            >
                            <span class="line-content">{line}</span>
                        </div>
                    {:else}
                        &nbsp;
                    {/each}
                </div>
            {/if}

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

    .json-tree {
        font-family: monospace;
        font-size: 15px;
        padding: 20px;
        border: 1px solid var(--border-color);
        border-radius: 4px;
        background-color: var(--textarea-bg-color);
        color: var(--text-color);
    }

    .container {
        max-width: 800px;
        margin: 20px auto;
        padding: 20px;
        border-radius: 8px;
        box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
        background-color: var(--bg-color);
    }

    .top-row {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 10px;
    }

    .top-row-item {
        padding: 5px;
    }

    .top-row-item button {
        width: 50px;
        font-size: 25px;
        padding: 0;
        padding-top: 5px;
        padding-bottom: 5px;
        line-height: 1.2;
    }

    .top-row-right {
        margin-left: auto;
    }

    .header {
        display: flex;
        text-align: center;
        align-items: center;
        justify-content: center;
        font-size: 20px;
        font-weight: bold;
        color: var(--text-color);
    }

    .source-file {
        text-align: center;
        font-size: 12px;
        font-weight: bold;
        padding-bottom: 10px;
    }

    .snippet {
        display: flex;
        flex-direction: column;
        font-family: monospace;
        font-size: 13px;
        padding: 10px;
        border-radius: 5px;
        overflow: auto;
        border: solid 1px var(--calm-color);
        margin-bottom: 10px;
    }

    .code-line {
        display: flex;
        padding: 2px 5px;
    }
    .mark {
        width: 30px;
        text-align: right;
        margin-right: 0;
    }
    .highlight {
        background-color: var(--snippet-highlight-color);
    }
    .line-number {
        width: 30px;
        text-align: right;
        margin-right: 10px;
        color: var(--calm-color);
    }
    .highlight .line-number {
        color: var(--text-color);
    }

    .line-content {
        white-space: pre;
    }
</style>
