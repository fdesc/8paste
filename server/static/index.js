let fileSelected = false;
const editor = document.getElementById('codeEditor');
const lineNumbers = document.getElementById('lineNumbers');
const editorContent = document.querySelector('.editor-content');

// backend related
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

// frontend related
document.getElementById('copyButton').addEventListener('click', copyToClipboard);
document.getElementById('pasteButton').addEventListener('click', pasteToEditor);
document.getElementById('uploadButton').addEventListener('click', openModal);
document.getElementById('downloadButton').addEventListener('click', downloadPasteValue);
document.getElementById('closeSmallModalButton').addEventListener('click', closeSmallModal);
document.getElementById('closeModalButton').addEventListener('click', closeModal);
document.getElementById('saveInfoButton').addEventListener('click', saveInfo);
document.getElementById('fileBtn').addEventListener('click', handleFileBtnClick);
document.getElementById('mainSelect').addEventListener('change', handleSelectChange);
document.getElementById('timeSelect').addEventListener('change', handleRadioChange);
document.getElementById('inputSelect').addEventListener('change', handleRadioChange);


// modal start
function initializeTimeSelectors() {
  const hourSelect = document.getElementById('hourSelect');
  const minuteSelect = document.getElementById('minuteSelect');
  const secondSelect = document.getElementById('secondSelect');

  for (let i = 0; i < 24; i++) {
    const option = document.createElement('option');
    option.value = i.toString().padStart(2, '0');
    option.textContent = i.toString().padStart(2, '0');
    hourSelect.appendChild(option);
  }

  for (let i = 0; i < 60; i++) {
    const minuteOption = document.createElement('option');
    minuteOption.value = i.toString().padStart(2, '0');
    minuteOption.textContent = i.toString().padStart(2, '0');
    minuteSelect.appendChild(minuteOption);

    const secondOption = document.createElement('option');
    secondOption.value = i.toString().padStart(2, '0');
    secondOption.textContent = i.toString().padStart(2, '0');
    secondSelect.appendChild(secondOption);
  }
}

function handleSelectChange() {
  const select = document.getElementById('mainSelect');
  const fileContainer = document.getElementById('fileUploadContainer');

  if (select.value === 'file') {
    fileContainer.classList.add('active');
  } else {
    fileContainer.classList.remove('active');
  }
}

function handleRadioChange() {
  const timeRadio = document.getElementById('timeSelect');
  const timeSelectors = document.getElementById('timeSelectors');

  if (timeRadio.checked) {
    timeSelectors.classList.add('active');
    inputContainer.classList.remove('active');
  } else {
    timeSelectors.classList.remove('active');
  }
}

function handleSealChange() {
  const radio = document.getElementById('sealYes');
  const inputContainer = document.getElementById('sealInputContainer');

  if (radio.checked) {
    inputContainer.classList.add('active');
  } else {
    inputContainer.classList.remove('active');
  }
}

document.getElementById('dialog').addEventListener('change', function(e) {
  const fileName = document.getElementById('fileName');
  const fileBtn = document.getElementById('fileBtn');
  if (e.target.files.length > 0) {
    fileName.textContent = e.target.files[0].name;
    fileBtn.firstChild.textContent = 'Remove';
    fileSelected = true;
  } else {
    fileName.textContent = 'No file chosen';
    fileBtn.firstChild.textContent = 'Choose File';
    fileSelected = false;
  }
});

async function saveInfo() {
  const mainSelect = document.getElementById('mainSelect');
  const selectedValue = mainSelect.value;
  const titleInput = document.querySelector('.form-group input.form-input[placeholder="Enter here..."]');
  const title = titleInput ? titleInput.value : '';

  const timeFormat = document.querySelector('input[name="timeFormat"]:checked').id;
  let duration = '';
  if (timeFormat === 'timeSelect') {
    const hour = document.getElementById('hourSelect').value;
    const minute = document.getElementById('minuteSelect').value;
    const second = document.getElementById('secondSelect').value;
    if (hour) duration += hour + 'h';
    if (minute) duration += minute + 'm';
    if (second) duration += second + 's';
  } else {
    duration = '';
  }

  const sealOption = document.querySelector('input[name="sealOption"]:checked').value;
  let password = '';
  if (sealOption === 'yes') {
    const passwordInput = document.querySelector('#sealInputContainer input.form-input');
    password = passwordInput ? passwordInput.value : '';
  } else {
    password = '';
  }

  let a = document.getElementById("footer-alert");
  if ((timeFormat === 'timeSelect') && (duration === 'hms' || duration === '')) {
    a.style.display = 'block';
    return;
  } else {
    a.style.display = 'none';
  }

  const formData = new FormData();
  let content = null;
  const info = {};
  info.title = title;
  if (password != '') {
    password = password
      .replace(/[<>'"&]/g, '')
      .replace(/[^\w\s\-_.@]/g, '')
      .trim()
      .substring(0, 70);

    info.sealed = true;
    formData.append('password',password);
  } else {
    info.sealed = undefined;
  }
  if (duration.length !== 0) {
    info.temporary = true;
    info.duration = duration;
  } else {
    info.temporary = false;
    info.duration = undefined;
  }
  if (selectedValue === 'text') {
    info.isfile = false;
    content = editor.value;
  } else if (selectedValue === 'file') {
    const file = document.getElementById('dialog').files[0]
    if (!file) {
      console.log('Invalid file');
      a.textContent = 'Invalid file';
      a.style.display = 'block';
      return;
    }
    info.isfile = true;
    content = file;
  }
  if (content === null) {
    console.log('Invalid content');
    a.textContent = 'Invalid file';
    a.style.display = 'block';
    return;
  }
  formData.append('content',content);
  formData.append('info',JSON.stringify(info));
  const response = await fetch(window.location.origin+"/upload", {
    method: 'POST',
    body: formData,
  });
  if (response.status !== 201) {
    a.textContent = 'Request failed';
    a.style.display = 'block';
    return;
  } else {
    const m = await response.json();
    a.style.display = 'none';
    closeModal();
    openSmallModal(window.location.origin+'/'+m.id);
  }

}

function handleFileBtnClick() {
  if (!fileSelected) {
    document.getElementById('dialog').click();
  } else {
    const fileInput = document.getElementById('dialog');
    fileInput.value = '';
    document.getElementById('fileName').textContent = 'No file chosen';
    document.getElementById('fileBtn').firstChild.textContent = 'Choose File';
    fileSelected = false;
  }
}

function resetModalFields() {
  document.querySelectorAll('.modal input[type="text"]').forEach(input => input.value = '');
  document.querySelectorAll('.modal input[type="password"]').forEach(input => input.value = '');
  const fileInput = document.getElementById('dialog');
  if (fileInput) fileInput.value = '';
  const fileName = document.getElementById('fileName');
  if (fileName) fileName.textContent = 'No file chosen';
  document.querySelectorAll('.modal select').forEach(select => select.selectedIndex = 0);
  const infiniteRadio = document.getElementById('inputSelect');
  if (infiniteRadio) infiniteRadio.checked = true;
  const sealNoRadio = document.getElementById('sealNo');
  if (sealNoRadio) sealNoRadio.checked = true;
  const footerAlert = document.getElementById('footer-alert');
  if (footerAlert) footerAlert.style.display = 'none';
}

function openModal() {
  const overlay = document.getElementById('modalOverlay');
  overlay.classList.add('active');
  document.body.style.overflow = 'hidden';
  resetModalFields();
}

function closeModal() {
  const overlay = document.getElementById('modalOverlay');
  overlay.classList.remove('active');
  document.body.style.overflow = 'auto';
  resetModalFields();
}

function openSmallModal(text) {
  document.getElementById('smallModal').style.display = 'block';
  document.getElementById('smallModalText').textContent = text;
  document.getElementById('smallModalOverlay').classList.add('active');
}

function closeSmallModal() {
  document.getElementById('smallModalOverlay').classList.remove('active');
}

document.getElementById('modalOverlay').addEventListener('click', function(e) {
  if (e.target === this) {
    closeModal();
  }
});

document.getElementById('smallModalOverlay').addEventListener('click', function(e) {
  if (e.target === this) {
    closeSmallModal();
  }
});

document.addEventListener('keydown', function(e) {
  if (e.key === 'Escape') {
    closeModal();
  }
});

window.addEventListener('load', function() {
  initializeTimeSelectors();
});
// modal end

// editor start
function updateLineNumbers() {
  const lines = editor.value.split('\n');
  const lineCount = lines.length;

  let numbersHtml = '';
  for (let i = 1; i <= lineCount; i++) {
    numbersHtml += i + '\n';
  }
  lineNumbers.textContent = numbersHtml.slice(0, -1);
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
// editor end
