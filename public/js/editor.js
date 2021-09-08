const editor = new EditorJS({
    holder: 'editorjs',
    autofocus: true,
    tools: {
        image: SimpleImage,
        quote: Quote
    },
    data: {
        "blocks" : [
            {
                "type" : "paragraph",
                "data" : {
                    "text" : "Hey. Meet the new Editor. On this page you can see it in action — try to edit this text."
                }
            },
        ]
    }
})

async function postData(data) {
    const response = await fetch('/post/my-post', {
        method: 'POST',
        body: data
    })
    if (response.status == 200) {
        window.location.replace("/post/my-post")
    }
}

String.prototype.escape = function() {
    var tagsToReplace = {
        '&': '&amp;',
        '<': '&lt;',
        '>': '&gt;'
    };
    return this.replace(/[&<>]/g, function(tag) {
        return tagsToReplace[tag] || tag;
    });
};

const formData = new FormData()

const titleForm = document.querySelector('#title-form')
titleForm.addEventListener('submit', (e) => {
    e.preventDefault()
    const input = e.target.querySelectorAll('.form-control')
    const file = input[0].files[0]
    const reader = new FileReader()
    reader.onloadend = () => document.querySelector('.masthead').style.backgroundImage = `url(${reader.result})`;

    if (file) {
        reader.readAsDataURL(file);
    }
    const postHeading = document.querySelector('.post-heading').children
    postHeading[0].innerHTML = input[1].value.escape()
    postHeading[1].innerHTML = input[2].value.escape()

    formData.append('title', input[1].value.escape())
    formData.append('subtitle', input[2].value.escape())
    formData.append('cover', input[0].files[0])

    titleForm.reset()
})

const publishBtn = document.querySelector('#publish')
publishBtn.addEventListener('click', async e => {
    const outputData = await editor.save()
    formData.append('data', JSON.stringify(outputData))
    postData(formData)
})