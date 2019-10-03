// Auto expand text fields

var autoExpand = function (field) {

	// Reset field height
	field.style.height = 'inherit';

	// Get the computed styles for the element
	var computed = window.getComputedStyle(field);

	// Calculate the height
	var height = parseInt(computed.getPropertyValue('border-top-width'), 10)
	             + parseInt(computed.getPropertyValue('padding-top'), 10)
	             + field.scrollHeight
	             + parseInt(computed.getPropertyValue('padding-bottom'), 10)
	             + parseInt(computed.getPropertyValue('border-bottom-width'), 10);

	field.style.height = height + 'px';

};

document.addEventListener('input', function (event) {
	if (event.target.tagName.toLowerCase() !== 'textarea') return;
	autoExpand(event.target);
}, false); 


// Add field when Issue is Billing Terminated

$(document).on('change','#issue-input',function(){
    var selection = $(this).val();
    switch(selection){
    case "1":
        $("#IIA").show()
        break;
    default:
        $("#IIA").hide()
        break;
    }
});

// Make fields focused

function focusParent(elem) {
    elem.parentElement.classList.add('is-focused');
}

// Adds text when value is selected

const issueSelector = document.getElementById("issue-input");
const commentTextField = document.getElementById("comment-input");

function addText() {
  const selectedIssue = issueSelector.options[issueSelector.selectedIndex].text;
  if (selectedIssue === "Legacy Contacts") {
      commentTextField.value = "Legacy -";
      focusParent(commentTextField);
  } else if (selectedIssue === "TNE Contacts") {
    commentTextField.value = "TNE -";
    focusParent(commentTextField);
}
}


