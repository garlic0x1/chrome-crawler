getLinks();

function absolutePath(href) {
	try {
		var link = document.createElement("a");
		link.href = href;
		return link.href;
	} catch (error) {}
}

function getLinks() {
	var array = []
	if (!document) return array;
	var allElements = document.querySelectorAll("*");
	for (var el of allElements) {
		if (el.href && typeof el.href ==='string') {
			array.push(absolutePath(el.href));
		}
	}
	return array;
}
