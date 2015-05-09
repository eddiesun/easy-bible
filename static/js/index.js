/**
 * Autocomplete
 */
$(function() {
	$('#searchButton').click(function(event) {
		event.preventDefault();
		autocomplete(event);
	});
	$('#searchInput').keyup(autocomplete);
	$('#searchInput').blur(function() {
		if (!$('#autocomplete').is(":hover")) {
			$('#autocomplete').hide();
		}
	});
	$('#autocomplete').on('mouseenter', 'a.autocompleteRow', function(event) {
		$('#autocompleteContainer .autocompleteRow.active').removeClass('active');
		$(this).addClass('active');
	});
});

var pendingAutocomplete
function autocomplete(event) {
	if (!isAutocompleteFiringKey(event.which)) {
		return
	}

	event.preventDefault();

	if (pendingAutocomplete != null) {
		pendingAutocomplete.abort();
		pendingAutocomplete = null;
	}

	var q = $('#searchInput').val()
	if (q && q.length >= 2) {
		if (!$('#autocomplete').is(':visible')) {
			$("#autocomplete").html("<img class='autocompleteLoading' src='/static/img/ajax-loader.gif' />");
		}
		$("#autocomplete").show();
		pendingAutocomplete = $.ajax({
			type: "GET",
			url: "/autocomplete",
			data: {query: q},
			success: function(responseBody) {
				$("#autocomplete").html(responseBody);
				highlightAutocompleteRow();
			}
		});
	} else {
		$("#autocomplete").hide();
		$("#autocomplete").html("");
	}
}

function highlightAutocompleteRow(index) {
	if (!index) {
		index = 0;
	}
	$('#autocompleteContainer .autocompleteRow').removeClass('active');
	$('#autocompleteContainer .autocompleteRow[data-index='+index+']').addClass('active');
}


/**
 * Partial verses loading
 */
$(function() {
	initPartial();
	$(window).resize(checkLoadingPartial);
	$('#btnLoadMore').click(reloadBook);
});
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

	loadPartial(true);
	$(window).scroll(function(){
		checkLoadingPartial();
	});
}

function checkLoadingPartial() {
	if (isLoadingMarkerInViewport()) {
		loadPartial();
	}
}

function isAutoPartialLoadingEnabled() {
	if (partialParams.v1 <= 0) {
		return true;
	}
	return false;
}

function loadPartial(firstLoad) {
	if (pendingPartial != null) {
		return;
	}

	if (calculateNextPartialParam()) {
		return;
	}

	if (!isAutoPartialLoadingEnabled() && !firstLoad) {
		$('#loadingMarker').hide();
		return;
	}

	pendingPartial = $.ajax({
		type: "GET",
		url: "/partial",
		data: partialParams,
		success: function(responseBody) {
			$("#partialContainer").append(responseBody);
			checkVerseFont();
			pendingPartial = null;
			checkLoadingPartial();
			if (firstLoad) {
				scrollToVerse();
				if (partialParams.v1 > 0) {
					$('#btnLoadMore').show();
				}
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
	}
	return noMorePartial;
}

function isLoadingMarkerInViewport() {
	var box = $('#loadingMarker')[0].getBoundingClientRect();
	return box.bottom - box.height < $(window).height();
}

function reloadBook(event) {
	var partialContainer = $('#partialContainer');
	var b = partialContainer.data('init-book-id'),
		c = partialContainer.data('init-chapter'),
		v1 = partialContainer.data('init-verse-begin'),
		v2 = partialContainer.data('init-verse-end');
	var param = {
		'b' : b,
		'c' : c
	};
	window.location.href = '?'+$.param(param)+'#'+c+'-'+v1;
}

function scrollToVerse() {
	var hash = location.hash.replace('#', '');
	var param = hash.split('-');
	if (param.length >= 2) {
		chapter = param[0];
		verse = param[1];
		var anchor = $("a[name='c="+chapter+"&v1="+verse+"']");
		if (anchor) {
			$('html, body').animate({scrollTop: anchor.offset().top-10}, 300, 'easeInSine');
			$(".partialTable tr[data-chapter="+chapter+"][data-verse="+verse+"]").effect("highlight", {}, 8000);
		}
	}
}





/**
 * menu
 */
var menuAnimating = false;
$(function() {
	$("button.menu").click(function() {
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
				$("button.menu").addClass('active');
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
				$("button.menu").removeClass('active');
			}
		}
	});
})

/**
 * Menu button
 */
$(function() {
	$('button.bookmenu').click(function() {
		// hide chapter menu if any are already visible
		$("#chapterBlock").hide();

		// hide book menu
		$('#bookBlock').hide();
		// show selected book menu
		displaySelectedBookMenu($(this).data('book-id'), $(this).html());

		// load chapter menu via Ajax. when done slide down chapter menu
		loadChapterButtons($(this).data('book-id'), function() {
			$("#chapterBlock").slideDown({
				duration: 300
			});
		});

		// scroll to top book menu if screen is too low
		scrollToBookMenu();

	});

	$('#unselectBookMenu').click(function() {
		$('#unselectBookMenu').hide();
		$('#selectedBookMenu').hide();
		$("#chapterBlock").hide();
		$('#bookBlock').slideDown(500);
	});
});

function loadChapterButtons(bookId, callback) {
	$.ajax({
		type: "GET",
		url: "/bookmenu",
		data: {b: bookId},
		success: function(responseBody) {
			$("#chapterBlock").html(responseBody);
			callback();
		}
	});
}

function displaySelectedBookMenu(bookId, bookName) {
	$('#unselectBookMenu').show();
	$('#selectedBookMenu').show();
	$('#selectedBookMenu').data('book-id', bookId).html(bookName);
}

function scrollToBookMenu() {
	var box = $('#menuContainer')[0].getBoundingClientRect();
	// if menu container is out of the screen, scroll to its back
	if (box.top < 0) {
		var body = $("html, body");
		body.animate({scrollTop:0}, '500', 'swing');
	}

}

/**
 * Hot keys
 */
var verseFontSize = 16;
$(function() {
	$(document).keypress(function(event) {
		if (event.which == 61) { // + key was pressed. increase verse font
			incVerseFont();
		}
		if (event.which == 45) { // - key was pressed. decrease verse font
			decVerseFont();
		}
	});

	$("#searchInput").keydown(function(event) {
		if (event.which == 38) { // up key
			event.preventDefault();
			moveAutocompleteRowUp();
		}
		if (event.which == 40) { // down key
			event.preventDefault();
			moveAutocompleteRowDown();
		}
		if (event.which == 13) { // enter key
			if ($('#autocomplete').is(':visible') && $('#autocompleteContainer .autocompleteRow.active')) {
				event.preventDefault();
				window.location.href = $('#autocompleteContainer .autocompleteRow.active').attr("href");
			}
		}
	});
});
function incVerseFont() {
	verseFontSize++;
	checkVerseFont();
}
function decVerseFont() {
	verseFontSize--;
	checkVerseFont();

	// if loading box appears due to decreasing font
	checkLoadingPartial();
}
function checkVerseFont() {
	$('.partialTable').css('font-size', verseFontSize+'px');
}

function isAutocompleteFiringKey(key) {
	if (key >= 48 && key <= 90) { // 0-9 a-z
		return true;
	}
	if (key >= 96 && key <= 105) { // number pad 0-9
		return true;
	}
	if (key == 46 || key == 8 || key == 32) { // delete key and backspace and space
		return true;
	}
	if (key == 1) { // mouse left button
		return true;
	}
	return false;
}

function moveAutocompleteRowDown() {
	var activeRow = $('#autocompleteContainer .autocompleteRow.active');
	if (activeRow) {
		var activeRowIndex = activeRow.data('index');
		activeRowIndex ++;
		if (activeRowIndex >= $('#autocompleteContainer .autocompleteRow').length) {
			activeRowIndex = $('#autocompleteContainer .autocompleteRow').length - 1;
		}
		activeRow.removeClass('active');
		$('#autocompleteContainer .autocompleteRow[data-index='+activeRowIndex+']').addClass('active');
	}
}
function moveAutocompleteRowUp() {
	var activeRow = $('#autocompleteContainer .autocompleteRow.active');
	if (activeRow) {
		var activeRowIndex = activeRow.data('index');
		activeRowIndex --;
		if (activeRowIndex < 0) {
			activeRowIndex = 0;
		}
		activeRow.removeClass('active');
		$('#autocompleteContainer .autocompleteRow[data-index='+activeRowIndex+']').addClass('active');
	}
}


