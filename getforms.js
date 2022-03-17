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
	var allElements = document.querySelectorAll("form");
	for (var el of allElements) {
		if (el.action && typeof el.action ==='string') {
			array.push(absolutePath(el.action));
		}	}
	return array;
}
