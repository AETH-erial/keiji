function insertHtmlAtCursor(textarea, html) {
    const start = textarea.selectionStart;
    const end = textarea.selectionEnd;
    const before = textarea.value.substring(0, start);
    const after = textarea.value.substring(end);

    textarea.value = before + html + after;

    // move cursor after inserted block
    textarea.selectionStart = textarea.selectionEnd = start + html.length;
}

function insertSelectedImage() {
    const select = document.getElementById("imageSelect");
    const imageId = select.value;

    if (!imageId) {
        alert("Please select an image first.");
        return;
    }

    const html = `<div class="text-center">
<img src="/api/v1/images/${imageId}" class="rounded img-fluid">
</div>`;

    const textarea = document.getElementById("myTextarea");

    insertHtmlAtCursor(textarea, html);
}


function toggleDropdown(el) {
    const options = el.parentElement.querySelector('.options');
    const open = options.style.display === "block";
    document.querySelectorAll('.image-select .options').forEach(o => o.style.display = "none");
    options.style.display = open ? "none" : "block";
}

function chooseImage(optionEl, id) {
    const container = optionEl.closest('.image-select');

    // set visible label
    container.querySelector('.selected span').textContent = id;

    // set real select value
    container.querySelector('.real-select').value = id;

    // close dropdown
    container.querySelector('.options').style.display = "none";
}

// close dropdown when clicking outside
document.addEventListener("click", function (e) {
    if (!e.target.closest(".image-select")) {
        document.querySelectorAll('.image-select .options').forEach(o => o.style.display = "none");
    }
});


function showPreview(src) {
    const box = document.getElementById('previewBox');
    box.innerHTML = `<img src="${src}">`;
    box.style.display = 'block';
}

function hidePreview() {
    document.getElementById('previewBox').style.display = 'none';
}
