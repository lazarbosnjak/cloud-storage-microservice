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
  mustChangePassword: boolean;
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

export interface ChangePasswordRequest {
  email: string;
  currentPassword: string;
  newPassword: string;
}

export interface ForgotPasswordRequest {
  email: string;
}

export interface ResetPasswordRequest {
  tokenId: string;
  newPassword: string;
  repeatedPassword: string;
}

export interface MagicLinkSignInResponse {
  accessToken: string;
  expiresInSeconds: number;
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

  changePassword(data: ChangePasswordRequest): Observable<unknown> {
    return this.http.post('/api/auth/change-password', data, {
      withCredentials: true,
    });
  }

  forgotPassword(data: ForgotPasswordRequest): Observable<unknown> {
    return this.http.post('/api/auth/forgot-password', data);
  }

  resetPassword(data: ResetPasswordRequest): Observable<unknown> {
    return this.http.post('/api/auth/reset-password', data);
  }

  logout(): Observable<unknown> {
    return this.http.post('/api/auth/logout', null, {
      withCredentials: true,
    });
  }

  requestMagicLink(data: ForgotPasswordRequest): Observable<unknown> {
    return this.http.post('/api/auth/magic-link', data);
  }

  magicLinkSignIn(tokenId: string): Observable<MagicLinkSignInResponse> {
    return this.http.post<MagicLinkSignInResponse>(
      '/api/auth/magic-link/sign-in',
      { tokenId },
      { withCredentials: true },
    );
  }
}
