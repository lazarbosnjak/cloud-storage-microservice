import { Component, OnInit, inject, input, signal } from '@angular/core';
import { HttpErrorResponse } from '@angular/common/http';
import { RouterLink } from '@angular/router';
import { firstValueFrom } from 'rxjs';

import { Auth } from '../../core/services/auth';
import { decodeJwtPayload, isTokenExpired } from '../../core/utils/jwt';

type Stage = 'loading' | 'done' | 'error';

@Component({
  selector: 'app-magic-link-sign-in',
  imports: [RouterLink],
  templateUrl: './magic-link-sign-in.html',
  styleUrl: './magic-link-sign-in.css',
})
export class MagicLinkSignIn implements OnInit {
  private readonly auth = inject(Auth);

  readonly tokenId = input.required<string>();

  readonly stage = signal<Stage>('loading');
  readonly errorMessage = signal('');
  readonly expiresInSeconds = signal<number | null>(null);

  async ngOnInit(): Promise<void> {
    try {
      const response = await firstValueFrom(
        this.auth.magicLinkSignIn(this.tokenId()),
      );

      const payload = decodeJwtPayload(response.accessToken);

      if (!response.accessToken || !payload || isTokenExpired(response.accessToken)) {
        this.fail('Invalid session token. Please sign in again.');
        return;
      }

      sessionStorage.setItem('access_token', response.accessToken);
      this.expiresInSeconds.set(response.expiresInSeconds);
      this.stage.set('done');
    } catch (error) {
      if (error instanceof HttpErrorResponse) {
        this.showError(error);
      } else {
        throw error;
      }
    }
  }

  private showError(error: HttpErrorResponse): void {
    const detail = (error.error as { detail?: string } | null)?.detail;
    this.fail(detail || 'Sign-in failed. Please try again.');
  }

  private fail(message: string): void {
    this.errorMessage.set(message);
    this.stage.set('error');
  }
}
