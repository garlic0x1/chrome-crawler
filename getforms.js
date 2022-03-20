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

	for (var f in array) {
		var form = JSON.parse(array[f]);
		var data = "";
		var hash = makeid(8);
		for (var i in form.inputs) {
			input = form.inputs[i];
			if (input.value.length > 0 || input.type == "hidden") {
				data += "&";
				data += input.name;
				data += "=";
				data += input.value;
			} else {
				data += "&";
				data += input.name;
				data += "=";
				data += hash;
			}
				
		}
		var http = new XMLHttpRequest();
		http.open('POST', form.url, true);
	
		//Send the proper header information along with the request
		http.setRequestHeader('Content-type', 'application/x-www-form-urlencoded');

		http.onreadystatechange = function() {//Call a function when the state changes.
    			if(http.readyState == 4 && http.status == 200) {
       				alert(http.responseText);
			}
		}
		http.send(data);
	}

	return array;
}
