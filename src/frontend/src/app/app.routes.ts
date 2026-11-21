
import { Routes } from '@angular/router';

export const routes: Routes = [
  {
    path: 'register',
    loadComponent: () =>
      import('./features/registration/registration')
        .then((m) => m.Registration),
  },
  {
    path: '',
    redirectTo: 'register',
    pathMatch: 'full',
  },
];
