import { Component, inject, signal, OnInit, computed } from '@angular/core';
import { DatePipe } from '@angular/common';
import { Router } from '@angular/router';
import { HttpErrorResponse } from '@angular/common/http';
import { firstValueFrom } from 'rxjs';

import { FileService } from '../../core/services/file';
import {
  FileMetadata,
  FolderMetadata,
  StorageMetadata,
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
  readonly successMessage = signal('');

  readonly storage = signal<StorageMetadata | null>(null);
  readonly storageError = signal('');
  readonly storagePercentage = computed(() => {
    const storage = this.storage();
    if (!storage || Number(storage.quota) <= 0) {
        return 0;
    } return Math.max( 0, Math.min( 100, (Number(storage.used ?? 0) / Number(storage.quota)) * 100,),
    );
  });

  readonly path = signal([
    { id: 'root', name: 'My Files' },
  ]);

  ngOnInit(): void {
    void this.loadContents(this.path());
    void this.loadStorage();
  }

  private checkSession(): boolean {
    const token = sessionStorage.getItem('access_token');

    if (!token || isTokenExpired(token)) {
      void this.router.navigate(['/sign-in']);
      return false;
    }

    return true;
  }

  private getParentId(): string {
    const current = this.path()[this.path().length - 1];

    return current.id === 'root' ? '' : current.id;
  }

  private showError(error: unknown): void {
    if (error instanceof HttpErrorResponse) {
      if (error.status === 401) {
        void this.router.navigate(['/sign-in']);
        return;
      }

      this.errorMessage.set(
        error.error?.detail || 'Request failed. Please try again.',
      );
    } else {
      this.errorMessage.set('An unexpected error occurred.');
    }
  }

  async loadContents(
    path: { id: string; name: string }[],
  ): Promise<void> {
    if (!this.checkSession()) {
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
      this.showError(error);
    } finally {
      this.loading.set(false);
    }
  }

  openFolder(folder: FolderMetadata): void {
    this.successMessage.set('');

    void this.loadContents([
      ...this.path(),
      { id: folder.id, name: folder.name },
    ]);
  }

  goBack(): void {
    if (this.path().length > 1) {
      this.successMessage.set('');
      void this.loadContents(this.path().slice(0, -1));
    }
  }

  refresh(): void {
    this.successMessage.set('');
    void this.loadContents(this.path());
    void this.loadStorage();
  }

  async createFolder(): Promise<void> {
    if (!this.checkSession()) {
      return;
    }

    const name = window.prompt('Enter folder name:')?.trim();

    if (!name) {
      return;
    }

    this.loading.set(true);
    this.errorMessage.set('');
    this.successMessage.set('');

    try {
      await firstValueFrom(
        this.fileService.createFolder(name, this.getParentId()),
      );

      await this.loadContents(this.path());

      this.successMessage.set('Folder created successfully.');
    } catch (error) {
      this.showError(error);
    } finally {
      this.loading.set(false);
    }
  }

  async renameFolder(folder: FolderMetadata): Promise<void> {
    if (!this.checkSession()) {
      return;
    }

    const name = window.prompt(
      'Enter new folder name:',
      folder.name,
    )?.trim();

    if (!name || name === folder.name) {
      return;
    }

    this.loading.set(true);
    this.errorMessage.set('');
    this.successMessage.set('');

    try {
      await firstValueFrom(
        this.fileService.renameFolder(folder.id, name),
      );

      await this.loadContents(this.path());

      this.successMessage.set('Folder renamed successfully.');
    } catch (error) {
      this.showError(error);
    } finally {
      this.loading.set(false);
    }
  }

  async renameFile(file: FileMetadata): Promise<void> {
    if (!this.checkSession()) {
      return;
    }

    const name = window.prompt(
      'Enter new file name:',
      file.name,
    )?.trim();

    if (!name || name === file.name) {
      return;
    }

    this.loading.set(true);
    this.errorMessage.set('');
    this.successMessage.set('');

    try {
      await firstValueFrom(
        this.fileService.renameFile(file.id, name),
      );

      await this.loadContents(this.path());

      this.successMessage.set('File renamed successfully.');
    } catch (error) {
      this.showError(error);
    } finally {
      this.loading.set(false);
    }
  }

  async uploadFile(event: Event): Promise<void> {
    const input = event.target as HTMLInputElement;
    const file = input.files?.[0];

    input.value = '';

    if (!file || !this.checkSession()) {
      return;
    }

    const existingFile = this.files().find(
      (item) => item.name === file.name,
    );

    if (
      existingFile &&
      !window.confirm('Upload a new version of this file?')
    ) {
      return;
    }

    this.loading.set(true);
    this.errorMessage.set('');
    this.successMessage.set('');

    try {
      await firstValueFrom(
        this.fileService.uploadFile(file, this.getParentId()),
      );

      await this.loadContents(this.path());
      await this.loadStorage();

      this.successMessage.set(
        existingFile
          ? 'New file version uploaded successfully.'
          : 'File metadata uploaded successfully.',
      );
    } catch (error) {
      this.showError(error);
    } finally {
      this.loading.set(false);
    }
  }

  async loadStorage(): Promise<void> {
  if (!this.checkSession()) {
    return;
  }

  this.storageError.set('');

  try {
    const response = await firstValueFrom(
      this.fileService.getStorage(),
    );

    this.storage.set(response);
  } catch (error) {
    this.storage.set(null);

    if (error instanceof HttpErrorResponse && error.status === 401) {
      void this.router.navigate(['/sign-in']);
      return;
    }

    this.storageError.set('Failed to load storage information.');
  }
}

formatBytes(value: string): string {
  const bytes = Number(value ?? 0);

  if (bytes >= 1024 ** 3) {
    return `${(bytes / 1024 ** 3).toFixed(2)} GB`;
  }

  if (bytes >= 1024 ** 2) {
    return `${(bytes / 1024 ** 2).toFixed(2)} MB`;
  }

  if (bytes >= 1024) {
    return `${(bytes / 1024).toFixed(2)} KB`;
  }

  return `${bytes} B`;
}
}