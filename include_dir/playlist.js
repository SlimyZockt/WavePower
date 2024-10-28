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

  await fetch("/moved", {
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
