function drag(e) {
  e.dataTransfer.setData("s_name", e.target.id);
}

function allowDrop(e) {
  e.preventDefault();
}
/**
 *@param {Number} id
 */
async function drop(e, id) {
  e.preventDefault();
  let data = e.dataTransfer.getData("s_name");

  console.table({
    grabbed: data,
    droped: id,
  });
  if (data === "" || id === "") {
    return;
  }

  await fetch("/api/moved", {
    method: "POST",
    body: JSON.stringify({
      grabbed: data,
      droped: id,
    }),
  });

  htmx.trigger("#playlist", "playlist-changed", {});

  //console.log("moved");
  //
}
