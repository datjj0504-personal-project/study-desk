import {
  CreateNote,
  CreateSubject,
  DeleteNote,
  GetNote,
  ListNotes,
  ListSubjects,
  UpdateNote,
} from "../../wailsjs/go/main/App";
import { notes as generatedNotes } from "../../wailsjs/go/models";

export interface Note {
  id: number;
  title: string;
  body: string;
  subjectId?: number;
  createdAt: string;
  updatedAt: string;
}

export interface NoteInput {
  title: string;
  body: string;
  subjectId?: number;
}

export interface Subject {
  id: number;
  name: string;
}

function toNote(note: generatedNotes.Note): Note {
  return {
    id: note.id,
    title: note.title,
    body: note.body,
    subjectId: note.subjectId ?? undefined,
    createdAt: String(note.createdAt),
    updatedAt: String(note.updatedAt),
  };
}

export async function listNotes(): Promise<Note[]> {
  return (await ListNotes()).map(toNote);
}

export async function getNote(id: number): Promise<Note> {
  return toNote(await GetNote(id));
}

export async function createNote(input: NoteInput): Promise<Note> {
  return toNote(
    await CreateNote(new generatedNotes.Note({
      title: input.title,
      body: input.body,
      subjectId: input.subjectId,
    })),
  );
}

export async function updateNote(note: Note): Promise<void> {
  await UpdateNote(new generatedNotes.Note(note));
}

export async function deleteNote(id: number): Promise<void> {
  await DeleteNote(id);
}

export async function listSubjects(): Promise<Subject[]> {
  return (await ListSubjects()).map((subject) => ({
    id: subject.id,
    name: subject.name,
  }));
}

export async function createSubject(name: string): Promise<Subject> {
  const subject = await CreateSubject(new generatedNotes.Subject({ name }));
  return { id: subject.id, name: subject.name };
}
