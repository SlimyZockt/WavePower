"use strict";

const PLAY_ICON = `<span class="iconify tabler--player-play w-6 h-6"></span>`;
const PAUSE_ICON = `<span class="iconify tabler--player-pause w-6 h-6"></span>`;

const VOLUME_ICON = `<span class="iconify tabler--volume-off w-6 h-6"></span>`;

const VOLUME_OFF_ICON = `<span class="iconify tabler--volume w-6 h-6"></span>`;

/**
 * @type {HTMLButtonElement}
 */
const PLAY_BTN = document.getElementById("play-btn");
PLAY_BTN.innerHTML = PLAY_ICON;

/**
 * @type {HTMLButtonElement}
 */
const MUTE_BTN = document.getElementById("mute-btn");
MUTE_BTN.innerHTML = VOLUME_OFF_ICON;

/**
 * @type {HTMLAudioElement}
 */
const AUDIO = document.getElementById("audio");
AUDIO.volume = 0.5;

/**
 * @type {HTMLInputElement}
 */
const TIME_LINE = document.getElementById("time-line");

/**
 * @type {HTMLInputElement}
 */
const VOLUME = document.getElementById("volume");
VOLUME.value = AUDIO.volume * 100;

/**
 * @type {HTMLDivElement}
 */
const TIME_LEFT = document.getElementById("time-left");
TIME_LEFT.innerText = "00:00 / 00:00";

/**
 * @type {HTMLButtonElement}
 */
const LOGIN = document.getElementById("login");

/**
 * @type {HTMLButtonElement}
 */
const FORWARD = document.getElementById("forward");

/**
 * @type {HTMLButtonElement}
 */
const BACKWARD = document.getElementById("backward");

LOGIN.onclick = () => {
  window.location.href = "/auth/google";
};

let is_playing = false;
let last_tracked_volume = AUDIO.volume;
let is_muted = false;
let curAudioID = -1;

/**
 * @type {HTMLDivElement | null}
 */
let pl_item_play_btn = null;

FORWARD.addEventListener("click", () => {
  AUDIO.currentTime = AUDIO.duration;
});

loadAudio("assets/lofi/output.m3u8", false);

function togglePlayer() {
  if (!is_playing) {
    PLAY_BTN.innerHTML = PAUSE_ICON;
    if (pl_item_play_btn !== null) {
      pl_item_play_btn.innerHTML = PAUSE_ICON;
    }
    AUDIO.play();
    is_playing = true;
    return;
  }

  PLAY_BTN.innerHTML = PLAY_ICON;
  if (pl_item_play_btn !== null) {
    pl_item_play_btn.innerHTML = PLAY_ICON;
  }
  AUDIO.pause();
  is_playing = false;
}

PLAY_BTN.addEventListener("click", togglePlayer);

/**
 * @param {Number} time_s
 * @returns {string}
 */
function displayTime(time_s) {
  const hours = Math.floor(time_s / 3600);
  time_s = time_s - hours * 3600;
  const minutes = Math.floor(time_s / 60);
  const seconds = Math.floor(time_s - minutes * 60);

  let str = hours > 0 ? `${hours}`.padStart(2, "0") + ":" : "";
  return (
    str + `${minutes}`.padStart(2, "0") + ":" + `${seconds}`.padStart(2, "0")
  );
}

let tl_pressed = false;
TIME_LINE.addEventListener("mousedown", () => {
  tl_pressed = true;
});

TIME_LINE.addEventListener("mouseup", (e) => {
  AUDIO.currentTime = AUDIO.duration * (e.target.value / 100);
  tl_pressed = false;
});

AUDIO.addEventListener("loadeddata", () => {
  TIME_LEFT.innerText = `${displayTime(AUDIO.currentTime)} / ${displayTime(AUDIO.duration)}`;
});

AUDIO.addEventListener("timeupdate", () => {
  if (Number.isNaN(AUDIO.duration)) return;
  if (tl_pressed === true) return;
  TIME_LINE.value = `${(AUDIO.currentTime / AUDIO.duration) * 100}`;
  TIME_LEFT.innerText = `${displayTime(AUDIO.currentTime)} / ${displayTime(AUDIO.duration)}`;
});

AUDIO.addEventListener("ended", () => {
  const src = `api/audio/${curAudioID + 1}/output.m3u8`;
  loadAudio(src, true, curAudioID + 1);
});

MUTE_BTN.addEventListener("click", () => {
  if (is_muted) {
    MUTE_BTN.innerHTML = VOLUME_OFF_ICON;
    AUDIO.volume = last_tracked_volume;
    VOLUME.value = last_tracked_volume * 100;
    TIME_LEFT.innerText = "00:00 / 00:00";
    is_muted = false;
    return;
  }

  MUTE_BTN.innerHTML = VOLUME_ICON;
  last_tracked_volume = AUDIO.volume;
  AUDIO.volume = 0.0;
  VOLUME.value = 0.0;
  is_muted = true;
});

VOLUME.addEventListener("input", (e) => {
  if (is_muted) {
    MUTE_BTN.innerHTML = VOLUME_OFF_ICON;
    is_muted = false;
  }

  AUDIO.volume = e.target.value / 100;
});

/**
 * @param {Event}
 */
// eslint-disable-next-line
async function fileuploud_change() {
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

  // eslint-disable-next-line
  htmx.trigger("#playlist", "playlist-changed", {});
}

/**
 * @param {string} src
 * @param {boolean} autoplay
 * @param {string} id
 */
async function loadAudio(src, autoplay, id) {
  curAudioID = id;

  // eslint-disable-next-line
  if (Hls.isSupported()) {
    // eslint-disable-next-line
    let hls = new Hls();
    hls.loadSource(src);
    hls.attachMedia(AUDIO);
    // eslint-disable-next-line
    hls.on(Hls.Events.MANIFEST_PARSED, function () {
      TIME_LINE.value = "0";
      AUDIO.currentTime = 0;
    });
  } else if (AUDIO.canPlayType("application/vnd.apple.mpegurl")) {
    AUDIO.src = src;
  }

  AUDIO.autoplay = autoplay;

  if (autoplay) {
    PLAY_BTN.innerHTML = PAUSE_ICON;
    is_playing = true;

    if (pl_item_play_btn !== null) {
      pl_item_play_btn.innerHTML = PLAY_ICON;
    }

    pl_item_play_btn = document.getElementById(`pl-p-${id}`);
    if (pl_item_play_btn !== null) {
      pl_item_play_btn.innerHTML = PAUSE_ICON;
    }
    return;
  }
  PLAY_BTN.innerHTML = PLAY_ICON;
  is_playing = false;
}

const TRACK_DISPLAY = document.getElementById("track-display");

/**
 * @param {string} id
 */
function onSelectAudio(id) {
  const src = `api/audio/${id}/output.m3u8`;

  fetch(`/api/track_display/${id}`, {
    method: "POST",
  })
    .catch(() => "")
    .then(async (res) => {
      TRACK_DISPLAY.innerHTML = await res.text();
    });

  loadAudio(src, true, id);
}

/**
 * @param {string} id
 */
// eslint-disable-next-line
function togglePlayerPLI(id) {
  if (id === curAudioID) {
    togglePlayer();
  } else {
    onSelectAudio(id);
  }
}
