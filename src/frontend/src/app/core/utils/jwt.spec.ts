import { decodeJwtPayload, isTokenExpired } from './jwt';

// payload: {"sub":"user-1","role":"USER","exp":1893456000} (2030-01-01)
const VALID_TOKEN =
  'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.' +
  'eyJzdWIiOiJ1c2VyLTEiLCJyb2xlIjoiVVNFUiIsImV4cCI6MTg5MzQ1NjAwMH0.' +
  'sig';

describe('decodeJwtPayload', () => {
  it('decodes the payload of a valid token', () => {
    const payload = decodeJwtPayload(VALID_TOKEN);
    expect(payload).toEqual({ sub: 'user-1', role: 'USER', exp: 1893456000 });
  });

  it('returns null for a token with wrong segment count', () => {
    expect(decodeJwtPayload('abc.def')).toBeNull();
  });

  it('returns null for non-JSON payload', () => {
    const token = 'aaa.bm90LWpzb24.sig'; // base64 of "not-json"
    expect(decodeJwtPayload(token)).toBeNull();
  });
});

describe('isTokenExpired', () => {
  it('returns false for a token expiring in 2030', () => {
    expect(isTokenExpired(VALID_TOKEN)).toBe(false);
  });

  it('returns true when there is no exp claim', () => {
    const token = 'aaa.e30.sig'; // base64 of "{}"
    expect(isTokenExpired(token)).toBe(true);
  });
});
