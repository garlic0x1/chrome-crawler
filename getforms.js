getForms();

function absolutePath(href) {
	try {
		var link = document.createElement("a");
		link.href = href;
		return link.href;
	} catch (error) {}
}

function getForms() {
	var array = []
	if (!document) return array;
	var allElements = document.querySelectorAll("form");
	for (var el of allElements) {
		var inputs = []; var allIns = el.querySelectorAll("input");
		for (var ch of allIns) {
			inputs.push({
					"type": ch.type,
					"name": ch.name,
					"value": ch.value,
			})
		}
		array.push(JSON.stringify({
			"url": absolutePath(el.action),
			"method": el.method,
			"inputs": inputs,	
		}));
	}	
	return array;
}
