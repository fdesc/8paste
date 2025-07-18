const textInput = document.getElementById('textInput');
const enterButton = document.getElementById('enterButton');
const output = document.getElementById('output');

function sanitizeInput(input) {
  return input
    .replace(/[<>'"&]/g, '')
    .replace(/[^\w\s\-_.@]/g, '')
    .trim()
    .substring(0, 70);
}

async function handleSubmit() {
  const rawInput = textInput.value;

  if (!rawInput.trim()) {
    return;
  }

  const sanitizedInput = sanitizeInput(rawInput);

  let link = window.location.href;
  const arr = link.split("/");
  const response = await fetch("/"+arr[arr.length-1], {
    method: 'POST',
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded',
    },
    body: 'password=' + encodeURIComponent(sanitizedInput)
  });
  textInput.value = '';
  if (response.ok) {
    sessionStorage.setItem('tmp',sanitizedInput);
    document.open();
    document.write(await response.text())
    document.close();
  } else if (response.status === 429) {
    textInput.placeholder = 'Too many requests!';
  } else {
    textInput.placeholder = 'Incorrect!';
  }
}

enterButton.addEventListener('click', handleSubmit);

textInput.addEventListener('keypress', function(e) {
  if (e.key === 'Enter') {
    handleSubmit();
  }
});

textInput.addEventListener('input', function() {
  enterButton.disabled = !textInput.value.trim();
});

enterButton.disabled = true;
