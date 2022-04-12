package main

var (
GetForms = `
getForms();

function absolutePath(href) {
	try {
		var link = document.createElement("a");
		link.href = href;
		return link.href;
	} catch (error) {}
}

function makeid(length) {
	var result           = '';
	var characters       = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz';
	var charactersLength = characters.length;
	for ( var i = 0; i < length; i++ ) {
		result += characters.charAt(Math.floor(Math.random() * charactersLength));
   	}
   	return result;
}

function getForms() {
	var array = [];
	if (!document) return array;
	var allElements = document.querySelectorAll("form");
	for (var el of allElements) {
		var inputs = []; var allIns = el.querySelectorAll("input");
		for (var ch of allIns) {
			inputs.push({
					"Type": ch.type,
					"Name": ch.name,
					"Value": ch.value,
			});
		}
		array.push({
			"URL": absolutePath(el.action),
			"Method": el.method,
			"Inputs": inputs,	
			"Hash": "",
			"Reflected": "false",
		});
	}

	for (var f in array) {
		form = array[f];
		var data = [];
		var hash = makeid(8);
		for (var i in form.Inputs) {
			input = form.Inputs[i];
			if (input.Value.length > 0 || input.Type == "hidden") {
				data.push("&");
				data.push(input.Name);
				data.push("=");
				data.push(input.Value);
			} else {
				data.push("&");
				data.push(input.Name);
				data.push("=");
				data.push(hash);
			}
				
		}
		console.log(data.join(''));
		var http = new XMLHttpRequest();
		http.open('POST', form.URL, true);
	
		//Send the proper header information along with the request
		http.setRequestHeader('Content-type', 'application/x-www-form-urlencoded');

		http.onreadystatechange = function() {//Call a function when the state changes.
    			if (http.readyState === http.DONE) {
       				if (http.responseText.includes(hash)) {
					form.Reflected = "true";
					alert(form.URL);
				}
    			}
		};
		form.Hash = hash;
		http.send(data.join(''));
	}

	return array;
}
`

GetLinks = `
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
`

GetScripts = `
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
`
)
