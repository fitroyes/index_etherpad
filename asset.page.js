document.querySelectorAll(".item").forEach((item) => {
	item.onclick = (_) =>
		document.querySelector("iframe").src = item.dataset.url;
});
