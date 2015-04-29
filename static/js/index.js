$(function() {
	$('#searchButton').click(autocomplete);
	$('#searchInput').keyup(autocomplete);
	loadPartial();
});

var pendingAutocomplete
function autocomplete(event) {
	event.preventDefault();

	if (pendingAutocomplete != null) {
		pendingAutocomplete.abort();
		pendingAutocomplete = null;
	}

	var q = $('#searchInput').val()
	if (q && q.length > 1) {
		pendingAutocomplete = $.ajax({
			type: "GET",
			url: "/autocomplete",
			data: {query: q},
			success: function(responseBody) {
				$("#autocomplete").html(responseBody)
			}
		});
	} else {
		$("#autocomplete").html("")
	}
}

var pendingPartial
function loadPartial() {
	if (pendingPartial != null) {
		pendingPartial.abort();
		pendingPartial = null;
	}

	pendingPartial = $.ajax({
		type: "GET",
		url: "/partial",
		data: {
			b: 1,
			c: 1,
			v1: 1,
			v2: 30
		},
		success: function(responseBody) {
			$("#partialContainer").append(responseBody);
		}
	});
}

$(window).scroll(function(){
	if ($(window).scrollTop() == $(document).height() - $(window).height()){
		console.log($(document).height(), $(window).height(), $(window).scrollTop());
		loadPartial();
	}
}); 
