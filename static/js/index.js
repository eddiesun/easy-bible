/**
 * Page load
 */
$(function() {
	$('#searchButton').click(autocomplete);
	$('#searchInput').keyup(autocomplete);
	initPartial();
});

/**
 * Autocomplete
 */
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


/**
 * Partial verses loading
 */
var partialParams = {
	b: 0,
	c: 0,
	v1: 0,
	v2: 0
};
var noMorePartial = false;
var pendingPartial = null;

function initPartial() {
	partialParams.b = $('#partialContainer').data('init-book-id');
	partialParams.c = $('#partialContainer').data('init-chapter');
	partialParams.v1 = $('#partialContainer').data('init-verse-begin');
	partialParams.v2 = $('#partialContainer').data('init-verse-end');

	if (isLoadingMarkerInViewport()) {
		loadPartial();
	}

	$(window).scroll(function(){
		if (isLoadingMarkerInViewport()) {
			loadPartial();
		}
	});
}

function loadPartial() {
	if (pendingPartial != null) {
		// console.log("loadPartial - nothing");
		return;
	}

	if (calculateNextPartialParam()) {
		console.log("loadPartial - No more partial");
		return;
	}

	console.log("loadPartial - !!");

	pendingPartial = $.ajax({
		type: "GET",
		url: "/partial",
		data: partialParams,
		success: function(responseBody) {
			$("#partialContainer").append(responseBody);
			pendingPartial = null;
			if (isLoadingMarkerInViewport()) {
				loadPartial();
			}
		}
	});
}

function calculateNextPartialParam() {
	var lastPartial = $('.partial').last();
	if (lastPartial.length > 0) {
		if (lastPartial.data('chapter') >= lastPartial.data('max-chapter')) {
			noMorePartial = true;
			$('#loadingMarker').hide();
		} else {
			noMorePartial = false;
			partialParams.c = lastPartial.data('chapter') + 1;
		}
		console.log("next partial param", partialParams);
	}
	return noMorePartial;
}

function isLoadingMarkerInViewport() {
	var box = $('#loadingMarker')[0].getBoundingClientRect();
	// console.log("In viewport? ", box.bottom - box.height < $(window).height());
	return box.bottom - box.height < $(window).height();
}



/**
 * menu
 */
var menuAnimating = false;
$(function() {
	$("button.menu").click(function() {
		console.log($(this), event);
		if (!menuAnimating) {
			if (!$('#menuContainer').is(':visible')) {
				menuAnimating = true;
				$("#menuContainer").slideDown({
					duration: 1000,
					easing: "easeOutBounce",
					complete: function() {
						menuAnimating = false;
					}
				});
			} else {
				if ($('#menuContainer').is(':visible')) {
					menuAnimating = true;
					$("#menuContainer").slideUp({
						duration: 1000,
						easing: "easeOutBounce",
						complete: function() {
							menuAnimating = false;
						}
					});
				}
			}
		}
	});
})
// function toggleMenu(event) {
// }
