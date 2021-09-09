let newData = {
    "blocks" : []
}

if (!data) {
    newData = {}
} else {
    for (let i=0; i<data.Blocks.length; i++) {
        if (data.Blocks[i].Type === "paragraph") {
            newData.blocks.push({
                "type": data.Blocks[i].Type,
                "data" : {
                    "text": data.Blocks[i].Data.Text
                }
            })
        } else {
            newData.blocks.push({
                "type": data.Blocks[i].Type,
                "data" : {
                    "file": data.Blocks[i].Data.File
                }
            })
        }
    }
}

const editor = new EditorJS({
    holder: 'editorjs',
    autofocus: true,
    tools: {
        image: SimpleImage,
        quote: Quote
    },
    data: newData
})

async function postData(data) {
    let method
    if (data) {
        method = 'PUT'
    } else {
        method = 'POST'
    }
    const response = await fetch('/post/my-post', {
        method: method,
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
    if (data) {
        formData.append('id', postId)
    }
    postData(formData)
})