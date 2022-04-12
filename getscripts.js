getScripts();

function absolutePath(href) {
	try {
		var link = document.createElement("a");
		link.href = href;
		return link.href;
	} catch (error) {}
}

function getScripts() {
	var array = []
	if (!document) return array;
	var allElements = document.querySelectorAll("script");
	for (var el of allElements) {
		if (el.src && typeof el.src === 'string') {
			array.push(absolutePath(el.src));
		}
	}
	return array
}
