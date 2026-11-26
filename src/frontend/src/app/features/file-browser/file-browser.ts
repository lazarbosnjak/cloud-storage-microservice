import { Component, inject, signal, OnInit } from '@angular/core';
import { DatePipe } from '@angular/common';
import { Router } from '@angular/router';
import { HttpErrorResponse } from '@angular/common/http';
import { firstValueFrom } from 'rxjs';

import { FileService } from '../../core/services/file';
import {
  FileMetadata,
  FolderMetadata,
} from '../../core/models/file-metadata';
import { isTokenExpired } from '../../core/utils/jwt';

@Component({
  selector: 'app-file-browser',
  imports: [DatePipe],
  templateUrl: './file-browser.html',
})
export class FileBrowser implements OnInit {
  private readonly fileService = inject(FileService);
  private readonly router = inject(Router);

  readonly files = signal<FileMetadata[]>([]);
  readonly folders = signal<FolderMetadata[]>([]);
  readonly loading = signal(false);
  readonly errorMessage = signal('');

  readonly path = signal([
    { id: 'root', name: 'My Files' },
  ]);

  ngOnInit(): void {
    void this.loadContents(this.path());
  }

  async loadContents(
    path: { id: string; name: string }[],
  ): Promise<void> {
    const token = sessionStorage.getItem('access_token');

    if (!token || isTokenExpired(token)) {
      this.router.navigate(['/sign-in']);
      return;
    }

    this.loading.set(true);
    this.errorMessage.set('');

    try {
      const current = path[path.length - 1];

      const response = await firstValueFrom(
        this.fileService.getFolderContents(current.id),
      );

      this.files.set(response.files ?? []);
      this.folders.set(response.folders ?? []);
      this.path.set(path);
    } catch (error) {
      if (error instanceof HttpErrorResponse && error.status === 401) {
        this.router.navigate(['/sign-in']);
      } else {
        this.errorMessage.set('Failed to load folder contents.');
      }
    } finally {
      this.loading.set(false);
    }
  }

  openFolder(folder: FolderMetadata): void {
    void this.loadContents([
      ...this.path(),
      { id: folder.id, name: folder.name },
    ]);
  }

  goBack(): void {
    if (this.path().length > 1) {
      void this.loadContents(this.path().slice(0, -1));
    }
  }
}