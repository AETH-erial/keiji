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
