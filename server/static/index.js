const editor = document.getElementById('codeEditor');
const lineNumbers = document.getElementById('lineNumbers');
const editorContent = document.querySelector('.editor-content');

async function getPasteInfo() {
  const arr = window.location.href.split("/");
  const formData = new FormData();
  formData.append("id",JSON.stringify({id:arr[arr.length-1]}));
  const response = await fetch(window.location.origin+"/info", {
    method: 'POST',
    body: formData,
  });
  return await response.json()
}

async function downloadPasteValue() {
  const arr = window.location.href.split("/");
  const formData = new FormData();
  formData.append("id",JSON.stringify({id:arr[arr.length-1]}));
  let info = await getPasteInfo();
  if (info.sealed) {
    let t = sessionStorage.getItem('tmp');
    formData.append("password",t);
    sessionStorage.removeItem('tmp');
  }
  const response = await fetch(window.location.origin+"/download", {
    method: 'POST',
    body: formData,
  });
  // https://stackoverflow.com/questions/63942715/how-to-download-a-readablestream-on-the-browser-that-has-been-returned-from-fetc
  const blob = await response.blob();
  const newblob = new Blob([blob]);
  const url = window.URL.createObjectURL(newblob);
  const a = document.createElement('a');
  a.href = url;
  a.setAttribute('download',info.title);
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  window.URL.revokeObjectURL(url);
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
