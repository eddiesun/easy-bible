
$(function() {
	$('#searchButton').click(autocomplete);
	$('#searchInput').keyup(autocomplete);
});

function autocomplete(event) {
	event.preventDefault();
	var q = $('#searchInput').val()
	if (q) {
		$("#autocomplete").load('/autocomplete?query='+q);
	} else {
		$("#autocomplete").html("")
	}
}