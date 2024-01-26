const articleElement = document.querySelector("article");
const loadingElement = document.querySelector("#loading");
const finishedElement = document.querySelector("#finished");
const canvasElement = document.querySelector("canvas");

const onWindowResize = () => {
  canvasElement.width = window.innerWidth;
  canvasElement.height = window.innerHeight;
};
window.addEventListener("resize", onWindowResize);
onWindowResize();

canvasElement.addEventListener("contextmenu", (event) => {
  event.preventDefault();
});

const hideLoading = () => {
  articleElement.style.display = "none";
  loadingElement.style.display = "none";
  canvasElement.style.display = "block";
};

const showFinished = () => {
  canvasElement.style.display = "none";
  articleElement.style.display = "block";
  finishedElement.style.display = "block";
};

console.log("Loading WebAssembly executable...");
const go = new Go();
const result = await WebAssembly.instantiateStreaming(
  fetch("web/main.wasm"),
  go.importObject
);
hideLoading();

console.log("Running WebAssembly executable...");
await go.run(result.instance);

console.log("Finished WebAssembly executable.");
showFinished();
