document.querySelectorAll(".item").forEach((item) => {
	document.querySelector(".selected")?.classList?.remove("selected");
	item.onclick = () => {
		item.classList.add("selected");
		document.querySelector("iframe").src = item.dataset.url;
	};
});
