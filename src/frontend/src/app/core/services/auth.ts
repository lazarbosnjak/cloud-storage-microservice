import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface RegisterRequest {
  email: string;
  password: string;
  repeatedPassword: string;
  firstName: string;
  lastName: string;
}

@Injectable({
  providedIn: 'root',
})
export class Auth {
  private readonly http = inject(HttpClient);

  register(data: RegisterRequest): Observable<unknown> {
    return this.http.post('/api/auth/register', data);
  }

  verifyRegistration(verificationToken: string): Observable<unknown> {
    const body = {
      verificationToken: verificationToken,
    };
    return this.http.post('/api/auth/register/verify', body);
  }
}
