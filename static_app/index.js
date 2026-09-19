let limit = Promise.withResolvers();
limit.resolve();

// Create an element, append the children, and return it.
// type is a string with the pattern: tag.class1.class2 ...
// Is tag is not present, use a simple div.
const $ = (type, obj, ...children) => {
		const [tagName, ...classNames] = type.split(/\.(\w+)/),
			element = document.createElement(tagName || "div");
		element.classList.add(...classNames.filter((x) => x));
		element.append(...children);
		Object.assign(element, obj || {});
		return element;
	},
	about = _about,
	fetchPad = async (host, id) => {
		const cachePath = `__txt/${host}/${id}`,
			response = await caches.match(cachePath);
		if (response) return response.text();

		const oldLimit = limit,
			newLimit = Promise.withResolvers();
		limit = newLimit;
		await oldLimit.promise;

		try {
			const pad = await fetch(`https://${host}/p/${id}/export/txt`);
			const text = await pad.text();
			(await caches.open("txt1")).put(
				cachePath,
				new Response(text, {
					headers: new Headers({
						"Content-Type": "text/plain;charset=utf-8",
						"Cache-Control": "max-age=3600",
					}),
				}),
			);
			return text;
		} catch (e) {
			console.error(e);
			return "";
		} finally {
			await setTimeout(newLimit.resolve, 6_000);
		}
	},
	getPads = (padContent) => {
		return padContent.match(/https:\/\/\S+\/p\/\S+/g).map((u) => {
			const uu = new URL(u);
			return [uu.hostname, uu.pathname.slice(3)];
		});
	},
	mainIndex = async (host, id) => {
		let mainBlock = $("main");
		const main = await fetchPad(host, id),
			pads = getPads(main),
			parseTitle = (text) =>
				text.split("\n")[0].replace(/^[\s#*=-]*/, "")
					.replace(/[\s#*=-]*$/, ""),
			renderPadsItem = ([host, id]) => {
				let text = "...";
				const name = $("span", 0, host, "/p/", id);
				fetchPad(host, id).then((data) => {
					text = data;
					name.replaceWith(parseTitle(text));
				});
				return $(
					"div.item",
					0,
					name,
					" ",
					$("a", {
						href: `https://${host}/p/${id}`,
						onclick(e) {
							e.preventDefault();
							const block = $("main", 0, text);
							mainBlock.replaceWith(block);
							mainBlock = block;
						},
					}, "[r]"),
					" ",
					$("a", {
						href: `https://${host}/p/${id}/export/txt`,
						onclick(e) {
							e.preventDefault();
							const iframe = $("iframe", {
								src: `https://${host}/p/${id}`,
							});
							mainBlock.replaceWith(iframe);
							mainBlock = iframe;
						},
					}, "[w]"),
				);
			};
		mainBlock.innerText = main;
		document.body.replaceChildren(
			$(
				"nav",
				0,
				$(
					"h1",
					0,
					document.title = parseTitle(main),
				),
				$("hr"),
				...pads.map(renderPadsItem),
				$("button", {
					async onclick() {
						const cache = await caches.open("txt1");
						for (const [host, id] of pads) {
							await cache.delete(`__txt/${host}/${id}`);
						}
					},
				}, "Remove cache"),
				$("hr"),
				about,
			),
			mainBlock,
		);
	},
	u = new URLSearchParams(location.search).get("u"),
	urlTarget = new URL(u ?? "");

u && mainIndex(urlTarget.hostname, urlTarget.pathname.slice(3));
