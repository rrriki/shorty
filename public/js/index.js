const endpoint = 'https://go-shorty.herokuapp.com'

document.addEventListener('DOMContentLoaded', function () {

    document.getElementById('url').addEventListener('click', () => {
        document.getElementById('url').value = ''
    });

    document.getElementById('shorten-button').addEventListener('click', () => {
        let longURL = document.getElementById('url').value;
        if (!longURL || longURL == '') {
            $('#modal').modal("show");
        }
        else { shorten(longURL); }

    });

    document.getElementById('expand-button').addEventListener('click', () => {
        let shortURL = document.getElementById('url').value;
        if (!shortURL || shortURL == '') {
            $('#modal').modal("show");
        } else { expand(shortURL); }
    });

    function shorten(url) {
        console.log('Shortening:', url)
        // Make a POST request to server
        fetch(`${endpoint}/shorten`, {
            headers: { "Content-Type": "application/json; charset=utf-8" },
            method: 'POST',
            body: JSON.stringify({ "longURL": url })
        })
            .then((res) => { return res.json() })
            .then((response) => {
                console.log('Success:', JSON.stringify(response))
                console.log(response.shortURL)
                document.getElementById('url').value = response.shortURL;
            }).catch((error) => { console.error('Error:', error) });
    }

    function expand(url) {
        console.log('Expanding:', url)
        // Make a POST request to server
        fetch(`${endpoint}/expand`, {
            headers: { "Content-Type": "application/json; charset=utf-8" },
            method: 'POST',
            body: JSON.stringify({ "shortURL": url })
        })
            .then((res) => { return res.json() })
            .then((response) => {
                console.log('Success:', JSON.stringify(response))
                document.getElementById('url').value = response.longURL;
            }).catch((error) => { console.error('Error:', error) });
    }

});


