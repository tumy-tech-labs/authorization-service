import json
from urllib import request


class Client:
    def __init__(self, base_url: str):
        self.base_url = base_url.rstrip('/')

    def _post(self, path: str, data: dict) -> tuple[int, str]:
        payload = json.dumps(data).encode('utf-8')
        req = request.Request(
            self.base_url + path,
            data=payload,
            headers={'Content-Type': 'application/json'},
            method='POST',
        )
        with request.urlopen(req) as resp:
            body = resp.read().decode('utf-8')
            return resp.status, body

    def check_access(self, tenant_id: str, subject: str, resource: str, action: str, conditions: dict | None = None) -> dict:
        status, body = self._post('/check-access', {
            'tenantID': tenant_id,
            'subject': subject,
            'resource': resource,
            'action': action,
            'conditions': conditions or {},
        })
        if status != 200:
            raise RuntimeError(f'unexpected status {status}: {body}')
        return json.loads(body)

    def compile_rule(self, tenant_id: str, rule: str) -> str:
        status, body = self._post('/compile', {'tenantID': tenant_id, 'rule': rule})
        if status != 200:
            raise RuntimeError(f'unexpected status {status}: {body}')
        return body

    def validate_policy(self, tenant_id: str, policy: str) -> str:
        status, body = self._post('/validate-policy', {'tenantID': tenant_id, 'policy': policy})
        if status != 200:
            raise RuntimeError(f'unexpected status {status}: {body}')
        return body

