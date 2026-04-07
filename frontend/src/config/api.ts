export const API_URL = import.meta.env.PUBLIC_API_URL || 'http://localhost:8080';

export interface Employee {
  id: string;
  dni: string;
  first_name: string;
  last_name: string;
  role: string;
  must_change_password: boolean;
}

export interface CheckAttendanceResponse {
  operation: 'check_in' | 'check_out';
  employee_id: string;
  date: string;
  check_in?: string;
  check_out?: string;
  message: string;
}

export interface LoginResponse {
  token: string;
  must_change_password: boolean;
  user: Employee;
}

// API Service for Fichar
export const api = {
  async checkAttendance(dni: string): Promise<CheckAttendanceResponse> {
    const response = await fetch(`${API_URL}/api/attendance/check`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ dni }),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ message: 'Error de conexión' }));
      throw new Error(error.message || 'Error al registrar asistencia');
    }

    return response.json();
  },

  async login(dni: string, password: string): Promise<LoginResponse> {
    const response = await fetch(`${API_URL}/api/auth/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ dni, password }),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ message: 'Credenciales inválidas' }));
      throw new Error(error.message || 'Error al iniciar sesión');
    }

    return response.json();
  },

  async changePassword(userId: string, newPassword: string, token: string): Promise<void> {
    const response = await fetch(`${API_URL}/api/auth/change-password`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
      },
      body: JSON.stringify({ user_id: userId, new_password: newPassword }),
    });

    if (!response.ok) {
      throw new Error('Error al cambiar contraseña');
    }
  },
};

// Local storage helpers for auth
export const auth = {
  getToken(): string | null {
    if (typeof window === 'undefined') return null;
    return localStorage.getItem('fichar_token');
  },

  setToken(token: string): void {
    if (typeof window === 'undefined') return;
    localStorage.setItem('fichar_token', token);
  },

  removeToken(): void {
    if (typeof window === 'undefined') return;
    localStorage.removeItem('fichar_token');
  },

  getUser(): Employee | null {
    if (typeof window === 'undefined') return null;
    const user = localStorage.getItem('fichar_user');
    return user ? JSON.parse(user) : null;
  },

  setUser(user: Employee): void {
    if (typeof window === 'undefined') return;
    localStorage.setItem('fichar_user', JSON.stringify(user));
  },

  removeUser(): void {
    if (typeof window === 'undefined') return;
    localStorage.removeItem('fichar_user');
  },

  isAuthenticated(): boolean {
    return !!this.getToken();
  },

  isAdmin(): boolean {
    const user = this.getUser();
    return user?.role === 'ADMIN';
  },
};