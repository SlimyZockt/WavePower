"use strict";

const PLAY_ICON = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 256 256"><path fill="currentColor" d="M232.4 114.49L88.32 26.35a16 16 0 0 0-16.2-.3A15.86 15.86 0 0 0 64 39.87v176.26A15.94 15.94 0 0 0 80 232a16.07 16.07 0 0 0 8.36-2.35l144.04-88.14a15.81 15.81 0 0 0 0-27ZM80 215.94V40l143.83 88Z"/></svg>`;
const PAUSE_ICON = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 256 256"><path fill="currentColor" d="M200 32h-40a16 16 0 0 0-16 16v160a16 16 0 0 0 16 16h40a16 16 0 0 0 16-16V48a16 16 0 0 0-16-16m0 176h-40V48h40ZM96 32H56a16 16 0 0 0-16 16v160a16 16 0 0 0 16 16h40a16 16 0 0 0 16-16V48a16 16 0 0 0-16-16m0 176H56V48h40Z"/></svg>`;

const UN_MUTE_ICON = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2 14.959V9.04C2 8.466 2.448 8 3 8h3.586a.98.98 0 0 0 .707-.305l3-3.388c.63-.656 1.707-.191 1.707.736v13.914c0 .934-1.09 1.395-1.716.726l-2.99-3.369A.98.98 0 0 0 6.578 16H3c-.552 0-1-.466-1-1.041M16 8.5c1.333 1.778 1.333 5.222 0 7M19 5c3.988 3.808 4.012 10.217 0 14"></path></svg>`;

const MUTE_ICON = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><g fill="none" stroke="currentColor" stroke-linecap="round" stroke-width="2"><path d="m22 15l-6-6m6 0l-6 6"></path><path stroke-linejoin="round" d="M2 14.959V9.04C2 8.466 2.448 8 3 8h3.586a.98.98 0 0 0 .707-.305l3-3.388c.63-.656 1.707-.191 1.707.736v13.914c0 .934-1.09 1.395-1.716.726l-2.99-3.369A.98.98 0 0 0 6.578 16H3c-.552 0-1-.466-1-1.041"></path></g></svg>`;

/**
 * @type {HTMLButtonElement}
 */
let play_btn = document.getElementById("play-btn");
play_btn.innerHTML = PLAY_ICON;

/**
 * @type {HTMLButtonElement}
 */
let mute_btn = document.getElementById("mute-btn");
mute_btn.innerHTML = UN_MUTE_ICON;

/**
 * @type {HTMLAudioElement}
 */
let audio = document.getElementById("audio");
audio.volume = 0.5;

/**
 * @type {HTMLInputElement}
 */
let time_line = document.getElementById("time-line");

/**
 * @type {HTMLInputElement}
 */
let volume = document.getElementById("volume");
volume.value = audio.volume * 100;

/**
 * @type {HTMLDivElement}
 */
let time_left = document.getElementById("time-left");
time_left.innerText = "00:00 / 00:00";

/**
 * @type {HTMLButtonElement}
 */
let login = document.getElementById("login");

login.onclick = async () => {
  window.location.href = "/auth/google";
  console.log(req);
};

let is_playing = false;
let last_tracked_volume = audio.volume;
let is_muted = false;

if (Hls.isSupported()) {
  let hls = new Hls();
  hls.loadSource("assets/lofi/output.m3u8");
  hls.attachMedia(audio);
  hls.on(Hls.Events.MANIFEST_PARSED, function () {});
}
// hls.js is not supported on platforms that do not have Media Source Extensions (MSE) enabled.
// When the browser has built-in HLS support (check using `canPlayType`), we can provide an HLS manifest (i.e. .m3u8 URL) directly to the video element throught the `src` property.
// This is using the built-in support of the plain video element, without using hls.js.
else if (audio.canPlayType("application/vnd.apple.mpegurl")) {
  audio.src = "assets/lofi/output.m3u8";
}

play_btn.addEventListener("click", () => {
  if (!is_playing) {
    play_btn.innerHTML = PAUSE_ICON;
    audio.play();
    is_playing = true;
    return;
  }

  play_btn.innerHTML = PLAY_ICON;
  audio.pause();
  is_playing = false;
});

/**
 * @param {Number} time_s
 * @returns {string}
 */
function display_time(time_s) {
  const hours = Math.floor(time_s / 3600);
  time_s = time_s - hours * 3600;
  const minutes = Math.floor(time_s / 60);
  const seconds = Math.floor(time_s - minutes * 60);

  let str = hours > 0 ? `${hours}`.padStart(2, "0") + ":" : "";
  return (
    str + `${minutes}`.padStart(2, "0") + ":" + `${seconds}`.padStart(2, "0")
  );
}

audio.addEventListener("loadedmetadata", () => {
  time_left.innerText = `${display_time(audio.currentTime)} / ${display_time(audio.duration)}`;
});

audio.addEventListener("timeupdate", () => {
  time_line.value = `${(audio.currentTime / audio.duration) * 100}`;
  time_left.innerText = `${display_time(audio.currentTime)} / ${display_time(audio.duration)}`;
});

mute_btn.addEventListener("click", () => {
  if (is_muted) {
    mute_btn.innerHTML = UN_MUTE_ICON;
    audio.volume = last_tracked_volume;
    volume.value = last_tracked_volume * 100;
    is_muted = false;
    return;
  }

  mute_btn.innerHTML = MUTE_ICON;
  last_tracked_volume = audio.volume;
  audio.volume = 0.0;
  volume.value = 0.0;
  is_muted = true;
});

volume.addEventListener("input", (e) => {
  if (is_muted) {
    mute_btn.innerHTML = UN_MUTE_ICON;
    is_muted = false;
  }

  audio.volume = e.target.value / 100;
});

time_line.addEventListener("input", (e) => {
  audio.currentTime = audio.duration * (e.target.value / 100);
});

//fileuploud.addEventListener("cancel", () => {});

/**
 * @param {Event}
 */
async function fileuploud_change(_) {
  /**
   * @type {HTMLInputElement} fileuploud
   */
  let fileuploud = document.getElementById("fileuploud");
  let files = fileuploud.files;

  if (files === null) return;

  console.log(files);
  let file = files[0];

  console.log(file.name);

  if (file === null) return;
  await fetch(`api/upload/${file.name}`, {
    body: file,
    method: "POST",
  });

  htmx.trigger("#playlist", "playlist-changed", {});
}
