#!/usr/bin/env -S deno run --allow-read=. --allow-write=. --unstable-bundle build.ts

const style = (await Deno.readTextFile("style.css"))
	.replaceAll(/[\n\t]+/g, "")
	.replaceAll(/\/\*.*?\*\//g, "")
	.replaceAll(": ", ":")
	.replaceAll(" {", "{")
	.replaceAll(";}", "}");

const results = await Deno.bundle({
	entrypoints: ["./index.js"],
	minify: true,
	write: false,
	platform: "browser",
});
for (const err of results.errors) {
	console.error(err);
}
for (const err of results.warnings) {
	console.warn(err);
}

Deno.writeTextFile(
	"output.html",
	(await Deno.readTextFile("main.html"))
		.replaceAll(/[\n\t]+/g, "")
		.replace("<!--STYLE-->", `<style>${style}</style>`)
		.replace(
			"<!--SCRIPT-->",
			`<script>${results.outputFiles?.[0].text() ?? ""}</script>`,
		),
);
