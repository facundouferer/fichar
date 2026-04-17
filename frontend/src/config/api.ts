// PUBLIC_API_URL is set at build time via .env file (PUBLIC_API_URL=http://backend:8080)
// Vite replaces this with the actual value during build
// If not set or running locally, default to localhost:8080
const envApiUrl = (import.meta as any).env?.PUBLIC_API_URL;

// For local development (localhost), use localhost. For production Docker, use the Docker network name.
export const API_URL = (typeof window !== 'undefined' && window.location.hostname === 'localhost') 
  ? 'http://localhost:8080' 
  : (envApiUrl || 'http://localhost:8080');

export interface Employee {
  id: string;
  dni: string;
  first_name: string;
  last_name: string;
  role: string;
  must_change_password: boolean;
  created_at?: string;
  updated_at?: string;
}

export interface Shift {
  id: string;
  name: string;
  start_time: string;
  end_time: string;
  expected_hours: number;
  created_at?: string;
}

export interface EmployeeShiftAssignment {
  id: string;
  employee_id: string;
  shift_id: string;
  start_date: string;
  end_date: string | null;
  employee_name?: string;
  shift_name?: string;
}

export interface MonthlySummary {
  year: number;
  month: number;
  employee_id: string;
  total_days: number;
  worked_days: number;
  missing_days: number;
  expected_hours: number;
  worked_hours: number;
  missing_hours: number;
  extra_hours: number;
  late_arrivals: number;
  daily_details: DailySummary[];
}

export interface DailySummary {
  date: string;
  check_in?: string;
  check_out?: string;
  worked_hours: number;
  expected_hours: number;
  is_late: boolean;
  shift_name?: string;
}

export interface DashboardSummary {
  total_employees: number;
  present_today: number;
  absent_today: number;
  late_arrivals_today: number;
  total_worked_hours: number;
  average_worked_hours: number;
  total_overtime_hours: number;
}

export interface AttendanceReport {
  employee_id: string;
  employee_name: string;
  dni: string;
  date: string;
  check_in?: string;
  check_out?: string;
  worked_hours: number;
  is_late: boolean;
  shift_name?: string;
}

export interface LateArrivalReport {
  employee_id: string;
  employee_name: string;
  dni: string;
  date: string;
  check_in: string;
  late_minutes: number;
  shift_name?: string;
}

export interface OvertimeReport {
  employee_id: string;
  employee_name: string;
  dni: string;
  date: string;
  check_in?: string;
  check_out?: string;
  worked_hours: number;
  overtime_hours: number;
  shift_name?: string;
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
  async checkAttendance(dni: string, latitude?: number, longitude?: number): Promise<CheckAttendanceResponse> {
    const body: Record<string, unknown> = { dni };
    
    if (latitude !== undefined && longitude !== undefined) {
      body.latitude = latitude;
      body.longitude = longitude;
    }
    
    const response = await fetch(`${API_URL}/api/attendance/check`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(body),
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

  // Admin API methods
  async getEmployees(token: string): Promise<Employee[]> {
    const response = await fetch(`${API_URL}/api/admin/employees`, {
      headers: { 'Authorization': `Bearer ${token}` },
    });
    if (!response.ok) throw new Error('Error al obtener empleados');
    const data = await response.json();
    return data.employees || [];
  },

  async createEmployee(token: string, employee: Partial<Employee>): Promise<Employee> {
    const response = await fetch(`${API_URL}/api/admin/employees`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
      },
      body: JSON.stringify(employee),
    });
    if (!response.ok) {
      const err = await response.json().catch(() => ({ message: 'Error al crear empleado' }));
      throw new Error(err.message);
    }
    return response.json();
  },

  async updateEmployee(token: string, id: string, employee: Partial<Employee>): Promise<Employee> {
    const response = await fetch(`${API_URL}/api/admin/employees/${id}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
      },
      body: JSON.stringify(employee),
    });
    if (!response.ok) {
      const err = await response.json().catch(() => ({ message: 'Error al actualizar empleado' }));
      throw new Error(err.message);
    }
    return response.json();
  },

  async deleteEmployee(token: string, id: string): Promise<void> {
    const response = await fetch(`${API_URL}/api/admin/employees/${id}`, {
      method: 'DELETE',
      headers: { 'Authorization': `Bearer ${token}` },
    });
    if (!response.ok) throw new Error('Error al eliminar empleado');
  },

  async getShifts(token: string): Promise<Shift[]> {
    const response = await fetch(`${API_URL}/api/admin/shifts`, {
      headers: { 'Authorization': `Bearer ${token}` },
    });
    if (!response.ok) throw new Error('Error al obtener turnos');
    const data = await response.json();
    return data.shifts || [];
  },

  async createShift(token: string, shift: Partial<Shift>): Promise<Shift> {
    const response = await fetch(`${API_URL}/api/admin/shifts`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
      },
      body: JSON.stringify(shift),
    });
    if (!response.ok) {
      const err = await response.json().catch(() => ({ message: 'Error al crear turno' }));
      throw new Error(err.message);
    }
    return response.json();
  },

  async updateShift(token: string, id: string, shift: Partial<Shift>): Promise<Shift> {
    const response = await fetch(`${API_URL}/api/admin/shifts/${id}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
      },
      body: JSON.stringify(shift),
    });
    if (!response.ok) {
      const err = await response.json().catch(() => ({ message: 'Error al actualizar turno' }));
      throw new Error(err.message);
    }
    return response.json();
  },

  async deleteShift(token: string, id: string): Promise<void> {
    const response = await fetch(`${API_URL}/api/admin/shifts/${id}`, {
      method: 'DELETE',
      headers: { 'Authorization': `Bearer ${token}` },
    });
    if (!response.ok) throw new Error('Error al eliminar turno');
  },

  async getEmployeeShifts(token: string, employeeId: string): Promise<EmployeeShiftAssignment[]> {
    const response = await fetch(`${API_URL}/api/admin/employees/${employeeId}/shifts`, {
      headers: { 'Authorization': `Bearer ${token}` },
    });
    if (!response.ok) throw new Error('Error al obtener asignaciones');
    const data = await response.json();
    return data.assignments || [];
  },

  async assignShift(token: string, assignment: { employee_id: string; shift_id: string; start_date: string; end_date?: string }): Promise<EmployeeShiftAssignment> {
    const response = await fetch(`${API_URL}/api/admin/employee-shifts`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
      },
      body: JSON.stringify(assignment),
    });
    if (!response.ok) {
      const err = await response.json().catch(() => ({ message: 'Error al asignar turno' }));
      throw new Error(err.message);
    }
    return response.json();
  },

  async getDashboardSummary(token: string): Promise<DashboardSummary> {
    const response = await fetch(`${API_URL}/api/reports/dashboard`, {
      headers: { 'Authorization': `Bearer ${token}` },
    });
    if (!response.ok) throw new Error('Error al obtener resumen');
    return response.json();
  },

  async getAttendanceReport(token: string, startDate: string, endDate: string, employeeId?: string): Promise<AttendanceReport[]> {
    let url = `${API_URL}/api/reports/attendance?start_date=${startDate}&end_date=${endDate}`;
    if (employeeId) url += `&employee_id=${employeeId}`;
    const response = await fetch(url, {
      headers: { 'Authorization': `Bearer ${token}` },
    });
    if (!response.ok) throw new Error('Error al obtener reporte de asistencia');
    return response.json();
  },

  async getMonthlyReport(token: string, employeeId: string, year: number, month: number): Promise<MonthlySummary> {
    const response = await fetch(
      `${API_URL}/api/reports/monthly?employee_id=${employeeId}&year=${year}&month=${month}`,
      { headers: { 'Authorization': `Bearer ${token}` } }
    );
    if (!response.ok) throw new Error('Error al obtener reporte mensual');
    return response.json();
  },

  async getLateArrivalsReport(token: string, startDate: string, endDate: string): Promise<LateArrivalReport[]> {
    const response = await fetch(
      `${API_URL}/api/reports/late-arrivals?start_date=${startDate}&end_date=${endDate}`,
      { headers: { 'Authorization': `Bearer ${token}` } }
    );
    if (!response.ok) throw new Error('Error al obtener reporte de tardanzas');
    return response.json();
  },

  async getOvertimeReport(token: string, startDate: string, endDate: string, employeeId?: string): Promise<OvertimeReport[]> {
    let url = `${API_URL}/api/reports/overtime?start_date=${startDate}&end_date=${endDate}`;
    if (employeeId) url += `&employee_id=${employeeId}`;
    const response = await fetch(url, {
      headers: { 'Authorization': `Bearer ${token}` },
    });
    if (!response.ok) throw new Error('Error al obtener reporte de horas extras');
    return response.json();
  },

  async generateSpecialReport(
    token: string,
    employeeId: string,
    header: string,
    customText: string,
    includeDays: boolean,
    includeHours: boolean,
    includeMonths: boolean,
    includePeriod: boolean,
    startDate?: string,
    endDate?: string
  ): Promise<Blob> {
    const response = await fetch(`${API_URL}/api/reports/special`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        employee_id: employeeId,
        header,
        custom_text: customText,
        include_days: includeDays,
        include_hours: includeHours,
        include_months: includeMonths,
        include_period: includePeriod,
        start_date: startDate || '',
        end_date: endDate || '',
      }),
    });
    if (!response.ok) {
      const error = await response.text();
      throw new Error(error || 'Error al generar informe especial');
    }
    return response.blob();
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