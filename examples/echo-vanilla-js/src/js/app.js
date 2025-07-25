// Note the ?url suffix
import wasmUrl from "@automerge/automerge/automerge.wasm?url";
// Note the `/slim` suffixes
import * as Automerge from "@automerge/automerge/slim";
import { Repo } from "@automerge/automerge-repo/slim";
import { IndexedDBStorageAdapter } from "@automerge/automerge-repo-storage-indexeddb";
import { BrowserWebSocketClientAdapter } from "@automerge/automerge-repo-network-websocket";

await Automerge.initializeWasm(wasmUrl);

const statusDiv = document.getElementById("status");
const documentPre = document.getElementById("document");
const updateButton = document.getElementById("updateButton");

// 1. Create a repo.
const repo = new Repo({
  storage: new IndexedDBStorageAdapter(),
  // Connect to the Go server's WebSocket endpoint.
  network: [new BrowserWebSocketClientAdapter("ws://localhost:1323/ws")],
  // Share policy is required to sync documents.
  sharePolicy: async (peerId, docId) => true,
});

// 2. Determine the document ID from the window's hash.
//    If no hash is present, generate a new UUID, set the hash, and reload.
let docId = window.location.hash.substring(1);

if (!docId) {
    docId = crypto.randomUUID();
    window.location.hash = docId;
}

const handle = repo.find(`automerge:${docId}`);

statusDiv.textContent = `Repo initialized. Document ID: ${docId}`;



// 3. Register a listener for document changes.
handle.on("change", ({ doc }) => {
  console.log("Document changed:", doc);
  // Update the UI with the new document state.
  documentPre.textContent = JSON.stringify(doc, null, 2);
  statusDiv.textContent = `Connected. Last change: ${new Date().toLocaleTimeString()}`;
});

// 4. Set up the button to modify the document.
updateButton.addEventListener("click", () => {
  if (!handle.isReady()) {
    alert("Document not ready yet. Please wait.");
    return;
  }
  // Use handle.change() to make a modification.
  handle.change((doc) => {
    doc.timestamp = new Date().toISOString();
    doc.counter = (doc.counter || 0) + 1;
    console.log("Making a change...");
  });
});

// Log the repo object for debugging.
console.log("Automerge Repo setup complete.", repo);
