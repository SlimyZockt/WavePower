"use strict";

let play_btn = document.getElementById("play_btn");
const play_icon = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 256 256"><path fill="currentColor" d="M232.4 114.49L88.32 26.35a16 16 0 0 0-16.2-.3A15.86 15.86 0 0 0 64 39.87v176.26A15.94 15.94 0 0 0 80 232a16.07 16.07 0 0 0 8.36-2.35l144.04-88.14a15.81 15.81 0 0 0 0-27ZM80 215.94V40l143.83 88Z"/></svg>`;
const pause_icon = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 256 256"><path fill="currentColor" d="M200 32h-40a16 16 0 0 0-16 16v160a16 16 0 0 0 16 16h40a16 16 0 0 0 16-16V48a16 16 0 0 0-16-16m0 176h-40V48h40ZM96 32H56a16 16 0 0 0-16 16v160a16 16 0 0 0 16 16h40a16 16 0 0 0 16-16V48a16 16 0 0 0-16-16m0 176H56V48h40Z"/></svg>`;

let is_playing = false;

// let audio_el = document.createElement("audio");
// document.appendChild(audio_el);

play_btn.innerHTML = play_icon;

/**
 * @type {AudioContext | undefined}
 */
let audioCtx = undefined;

/**
 * @type {AudioBuffer[]}
 */
let audio_buffers = [];

async function play_sound() {
  while (is_playing && audio_buffers.length > 0) {
    let source = audioCtx.createBufferSource();
    source.buffer = audio_buffers.shift();
    console.log(source.buffer.duration);
    console.log(source.buffer.length);
    source.connect(audioCtx.destination);
    source.start(0);
  }
}

play_btn.addEventListener("click", async () => {
  is_playing = !is_playing;
  play_btn.innerHTML = is_playing ? pause_icon : play_icon;

  if (audioCtx == undefined) {
    audioCtx = new AudioContext();
    audioCtx.createGain();
  }

  if (is_playing == false) {
    audioCtx.suspend();
    return;
  }

  audioCtx.resume();

  let res = await fetch("/audio/1", {
    method: "POST",
  });

  if (res.body == null) return;

  for await (const chunk of res.body) {
    if (chunk instanceof Uint8Array) {
      let audio_buffer = await audioCtx.decodeAudioData(chunk.buffer);
      audio_buffers.push(audio_buffer);
    }
  }

  play_sound();
});
