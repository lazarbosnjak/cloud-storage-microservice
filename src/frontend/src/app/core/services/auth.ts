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

export interface SignInRequest {
  email: string;
  password: string;
}

export interface SignInResponse {
  accessToken: string;
  expiresInSeconds: number;
  otpRequired: boolean;
  otpChallenge: string;
}

export interface VerifyOtpRequest {
  otpChallenge: string;
  code: string;
}

export interface VerifyOtpResponse {
  accessToken: string;
  expiresInSeconds: number;
  mustChangePassword: boolean;
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
  
  signIn(data: SignInRequest): Observable<SignInResponse> {
    return this.http.post<SignInResponse>('/api/auth/login', data, {
      withCredentials: true,
    });
  }

  verifyOtp(data: VerifyOtpRequest): Observable<VerifyOtpResponse> {
    return this.http.post<VerifyOtpResponse>('/api/auth/otp/verify', data, {
      withCredentials: true,
    });
  }
}
