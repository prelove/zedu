import { afterEach, describe, expect, it, vi } from 'vitest'
import { createBackup, getDashboard } from '../src/api/dashboard'
function response(data: unknown): Response { return { ok:true,status:200,json:async()=>({code:0,data}) } as Response }
describe('dashboard API adapters',()=>{const original=globalThis.fetch;afterEach(()=>{globalThis.fetch=original;vi.restoreAllMocks()});it('uses protected dashboard and backup paths',async()=>{const calls:string[]=[];globalThis.fetch=vi.fn().mockImplementation((url:string)=>{calls.push(url);return Promise.resolve(response({}))});await getDashboard('token');await createBackup('token');expect(calls).toEqual(['/dashboard','/system/backups'])})})
