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

const TRACK_DISPLAY = document.getElementById("track-display");
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

/**
 * @type {HTMLInputElement}
 */
const SHUFFLE = document.getElementById("shuffle");

/**
 * @type {HTMLInputElement}
 */
const REPEAT = document.getElementById("repeat");

LOGIN.onclick = () => {
  window.location.href = "/auth/google";
};

let is_playing = false;
let last_tracked_volume = AUDIO.volume;
let is_muted = false;
let curAudioID = -1;
let shuffle = false;
let repeat = false;

/**
 * @type {HTMLDivElement | null}
 */
let pl_item_play_btn = null;

FORWARD.addEventListener("click", () => {
  AUDIO.currentTime = AUDIO.duration;
});

BACKWARD.addEventListener("click", () => {
  AUDIO.currentTime = 0.0;
  const src = `api/audio/${curAudioID}/output.m3u8`;
  loadAudio(src, curAudioID);
});

BACKWARD.addEventListener("dblclick", async () => {
  const res = await fetch(`api/track/previous/${curAudioID}`, {
    method: "POST",
  });
  const id = await res.text();
  const src = `api/audio/${id}/output.m3u8`;
  loadAudio(src, id);
});

SHUFFLE.addEventListener("click", () => {
  shuffle = !shuffle;
});

REPEAT.addEventListener("click", () => {
  repeat = !repeat;
});

// first load audio
loadAudio("assets/lofi/output.m3u8", undefined);

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

AUDIO.addEventListener("ended", async () => {
  if (repeat) {
    const src = `api/audio/${curAudioID}/output.m3u8`;
    loadAudio(src, true, curAudioID);
    return;
  }

  let path = "/api/track/shuffle";

  if (!shuffle) {
    path = `/api/track/next/${curAudioID}`;
  }

  const res = await fetch(path, {
    method: "POST",
  });
  const id = await res.text();
  const src = `api/audio/${id}/output.m3u8`;
  loadAudio(src, id);
  return;
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
  const fileuploud = document.getElementById("fileuploud");
  const files = fileuploud.files;

  if (files === null) return;
  const file = files[0];

  const totalBytes = file.size * 1.5;
  let bytesUploaded = 0;

  const uploadProgress = document.getElementById("upload-progress");

  const progessTrackingStream = new TransformStream({
    transform(chunk, controller) {
      controller.enqueue(chunk);
      bytesUploaded += chunk.byteLength;
      console.log("upload progress:", bytesUploaded / totalBytes);
      uploadProgress.value = bytesUploaded / totalBytes;
    },
    flush() {
      console.log("completed stream");
    },
  });

  if (file === null) return;
  fetch(`api/upload/${file.name}`, {
    body: file.stream().pipeThrough(progessTrackingStream),
    method: "POST",
    duplex: "half",
  }).then(async (res) => {
    const err = await res.text();
    if (err !== "") {
      const uploadProgressTT = document.getElementById("upload-progress-tt");
      uploadProgressTT.setAttribute("data-tip", err);
      uploadProgressTT.className = "tooltip tooltip-open tooltip-error";
      setTimeout(() => {
        uploadProgressTT.setAttribute("data-tip", "");
        uploadProgressTT.className = "";
      }, 2000);
    }

    bytesUploaded = file.size * 1.5;
    uploadProgress.value = bytesUploaded / totalBytes;

    // eslint-disable-next-line
    htmx.trigger("#playlist", "playlist-changed", {});

    setTimeout(() => {
      uploadProgress.value = "0";
    }, 1000);
  });
}

/**
 * @param {string} src
 * @param {string | undefined} id
 */
async function loadAudio(src, id) {
  curAudioID = id;

  // eslint-disable-next-line
  if (Hls.isSupported()) {
    // eslint-disable-next-line
    let hls = new Hls();
    hls.loadSource(src);
    hls.attachMedia(AUDIO);
    // eslint-disable-next-line
    hls.on(Hls.Events.MANIFEST_PARSED, function() {
      TIME_LINE.value = "0";
      AUDIO.currentTime = 0;
    });
  } else if (AUDIO.canPlayType("application/vnd.apple.mpegurl")) {
    AUDIO.src = src;
  }

  if (id !== undefined) {
    const res = await fetch(`/api/track_display/${id}`, {
      method: "POST",
    });
    TRACK_DISPLAY.innerHTML = await res.text();

    AUDIO.autoplay = true;

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
  AUDIO.autoplay = false;
  is_playing = false;
}

/**
 * @param {string} id
 */
async function onSelectAudio(id) {
  const src = `api/audio/${id}/output.m3u8`;
  loadAudio(src, id);
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

// eslint-disable-next-line
function drag(e) {
  e.dataTransfer.setData("s_name", e.target.id);
}

// eslint-disable-next-line
function allowDrop(e) {
  e.preventDefault();
}
/**
 *@param {string} id
 */
// eslint-disable-next-line
async function drop(e, id) {
  e.preventDefault();
  /**
   *@type {string}
   */
  let data = e.dataTransfer.getData("s_name");

  console.table({
    grabbed: data.substring(3),
    droped: id.substring(3),
  });
  if (data === "" || id === "") {
    return;
  }

  await fetch("/api/moved", {
    method: "POST",
    body: JSON.stringify({
      grabbed: data.substring(3),
      droped: id.substring(3),
    }),
  });

  // eslint-disable-next-line
  htmx.trigger("#playlist", "playlist-changed", {});
}

