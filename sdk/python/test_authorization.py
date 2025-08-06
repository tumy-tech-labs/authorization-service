import json
import threading
import unittest
from http.server import BaseHTTPRequestHandler, HTTPServer

from authorization import Client


class Handler(BaseHTTPRequestHandler):
    def do_POST(self):
        if self.path == '/check-access':
            self.send_response(200)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            self.wfile.write(b'{"allow": true, "policyID": "p1", "reason": "ok"}')
        elif self.path == '/compile':
            self.send_response(200)
            self.end_headers()
            self.wfile.write(b'policy: allow')
        elif self.path == '/validate-policy':
            self.send_response(200)
            self.end_headers()
            self.wfile.write(b'policy is valid')
        else:
            self.send_response(404)
            self.end_headers()


class TestClient(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.server = HTTPServer(('localhost', 0), Handler)
        cls.port = cls.server.server_address[1]
        cls.thread = threading.Thread(target=cls.server.serve_forever)
        cls.thread.setDaemon(True)
        cls.thread.start()

    @classmethod
    def tearDownClass(cls):
        cls.server.shutdown()
        cls.thread.join()

    def setUp(self):
        self.client = Client(f'http://localhost:{self.port}')

    def test_check_access(self):
        decision = self.client.check_access('t', 's', 'r', 'a')
        self.assertTrue(decision['allow'])

    def test_compile_rule(self):
        yaml = self.client.compile_rule('t', 'rule')
        self.assertIn('policy', yaml)

    def test_validate_policy(self):
        resp = self.client.validate_policy('t', 'policy')
        self.assertIn('valid', resp)


if __name__ == '__main__':
    unittest.main()

