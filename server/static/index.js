const editor = document.getElementById('codeEditor');
const lineNumbers = document.getElementById('lineNumbers');
const editorContent = document.querySelector('.editor-content');

async function downloadPasteValue() {
  let currentlink = window.location.href;
  const arr = currentlink.split("/");
  const text = editor.value;
  const filename = 'paste-'+arr[arr.length-1];
  const blob = new Blob([text], { type: 'text/plain' });
  const link = document.createElement('a');
  link.href = URL.createObjectURL(blob);
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(link.href);
}

function updateLineNumbers() {
  const lines = editor.value.split('\n');
  const lineCount = lines.length;

  let numbersHtml = '';
  for (let i = 1; i <= lineCount; i++) {
    numbersHtml += i + '\n';
  }
  lineNumbers.textContent = numbersHtml.slice(0, -1);

  /*setTimeout(() => {
    const editorHeight = editor.scrollHeight;
    lineNumbers.style.height = editorHeight + 'px';
  }, 0);*/
}

function syncScroll() {
  lineNumbers.scrollTop = editor.scrollTop;
}

function copyToClipboard() {
  editor.setSelectionRange(0,99999);
  navigator.clipboard.writeText(editor.value);
}

function pasteToEditor() {
  navigator.clipboard.readText().then((pasted) => (editor.value += pasted));
}

editor.addEventListener('input', function() {
  updateLineNumbers();
});

editor.addEventListener('scroll', function() {
  syncScroll();
});

editor.addEventListener('keydown', function(e) {
  if (e.key === 'Tab') {
    e.preventDefault();
    const start = this.selectionStart;
    const end = this.selectionEnd;

    this.value = this.value.substring(0, start) + '    ' + this.value.substring(end);
    this.selectionStart = this.selectionEnd = start + 4;

    updateLineNumbers();
  }
});

editorContent.addEventListener('scroll', syncScroll);

updateLineNumbers();
