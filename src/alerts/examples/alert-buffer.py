#!/usr/bin/env python
from BaseHTTPServer import BaseHTTPRequestHandler, HTTPServer

class AlertHandler(BaseHTTPRequestHandler):
    def do_POST(self):
        # print out new alert changes
        print self.rfile.read(int(self.headers['Content-Length']))

        self.send_response(200)
        self.end_headers()


def run():
    httpd = HTTPServer(('0.0.0.0', 9099), AlertHandler)
    print 'Starting httpd...'
    httpd.serve_forever()

if __name__ == "__main__":
    run()
