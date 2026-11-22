// @project
import { AuthRole } from '@/enum';

/***************************  DATA - USERS  ***************************/

const mockUsers = [
  {
    id: '1',
    email: 'adryan@midaz.io',
    role: AuthRole.SUPER_ADMIN,
    contact: '+55 11 99999-9999',
    dialcode: '+55',
    firstname: 'Adryan',
    lastname: 'Alencar',
    password: 'Admin@123',
    access_token: 'midaz_admin_access_token'
  },
  {
    id: '2',
    email: 'financeiro@midaz.io',
    role: AuthRole.ADMIN,
    contact: '+55 11 98888-8888',
    dialcode: '+55',
    firstname: 'Operador',
    lastname: 'Financeiro',
    password: 'Admin@123',
    access_token: 'midaz_finance_access_token'
  },
  {
    id: '3',
    email: 'auditoria@midaz.io',
    role: AuthRole.USER,
    contact: '+55 11 97777-7777',
    dialcode: '+55',
    firstname: 'Auditor',
    lastname: 'Contábil',
    password: 'User@123',
    access_token: 'midaz_auditor_access_token'
  }
];

export default mockUsers;
