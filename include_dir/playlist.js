function drag(e) {
  e.dataTransfer.setData("s_name", e.target.id);
}

function allowDrop(e) {
  e.preventDefault();
}
/**
 *@param {DragEvent} e
 */
async function drop(e) {
  e.preventDefault();
  let data = e.dataTransfer.getData("s_name");

  if (data === "" || e.target.id === "") {
    return;
  }

  console.table({
    grabbed: data,
    droped: e.target.id,
  });

  await fetch("/api/moved", {
    method: "POST",
    body: JSON.stringify({
      grabbed: data,
      droped: e.target.id,
    }),
  });

  htmx.trigger("#playlist", "playlist-changed", {});

  //console.log("moved");
  //
}

/**
 * @param {Event} e
 */
async function onSelectSong(id) {
  if (Hls.isSupported()) {
    let hls = new Hls();
    hls.loadSource(`api/audio/${id}`);
    hls.attachMedia(audio);
    hls.on(Hls.Events.MANIFEST_PARSED, function () {});
  } else if (audio.canPlayType("application/vnd.apple.mpegurl")) {
    audio.src = `api/audio/${id}`;
  }
}
