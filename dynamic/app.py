"""Test sendint messages with mailgun."""

from aiohttp import web
import aiohttp
import os

URL = 'https://api.mailgun.net/v3/mg.highheathcattery.co.uk/messages'
AUTH = aiohttp.BasicAuth("api", os.environ['MAILGUN_TOKEN'])


async def contact(request):
    """Handler for the contact form."""
    async with aiohttp.ClientSession() as session:
        data = {
            "from": "High Heath Farm Cattery <info@highheathcattery.co.uk>",
            "to": ["smirlie@gmail.com"],
            "subject": "Hello",
            "text": "Testing some Mailgun awesomness!",
        }
        if request.query.get('debug'):
            data['o:testmode'] = 'true'
        async with session.post( URL, auth=AUTH, data=data) as response:
            json = await response.json()
    return web.json_response(json)


async def booking(request):
    """Handler for the booking form."""
    # async with aiohttp.ClientSession() as session:
    #     data = {
    #         "from": "High Heath Farm Cattery <info@highheathcattery.co.uk>",
    #         "to": ["smirlie@gmail.com"],
    #         "subject": "Hello",
    #         "text": "Testing some Mailgun awesomness!",
    #     }
    #     if request.query.get('debug'):
    #         data['o:testmode'] = 'true'
    #     async with session.post( URL, auth=AUTH, data=data) as response:
    #         json = await response.json()
    return web.json_response({"booking": True})


async def healthz(request):
    """A healthz endpoint for readiness."""
    return web.json_response({"healthy": True})


if __name__ == "__main__": 
    app = web.Application()
    app.add_routes([
        web.get('/healthz', healthz),
        web.post('/contact', contact),
        web.post('/booking', booking),
    ])
    web.run_app(app)
