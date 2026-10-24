import { afterEach, describe, expect, it, vi } from "vitest";
import { createNote, deleteNote, getNote, listNotes, updateNote } from "./notes";

const { bindings } = vi.hoisted(() => ({
  bindings: {
    CreateNote: vi.fn(),
    DeleteNote: vi.fn(),
    GetNote: vi.fn(),
    ListNotes: vi.fn(),
    UpdateNote: vi.fn(),
  },
}));

vi.mock("../../wailsjs/go/main/App", () => bindings);
vi.mock("../../wailsjs/go/models", () => ({
  notes: {
    Note: class {
      constructor(public source: Record<string, unknown>) {}
    },
  },
}));

afterEach(() => {
  vi.resetAllMocks();
});

describe("notes service", () => {
  it("maps list binding results to frontend notes", async () => {
    bindings.ListNotes.mockResolvedValue([
      { id: 2, title: "Note", body: "Body", createdAt: "created", updatedAt: "updated" },
    ]);

    await expect(listNotes()).resolves.toEqual([
      { id: 2, title: "Note", body: "Body", createdAt: "created", updatedAt: "updated" },
    ]);
  });

  it("wraps create, get, update, and delete bindings", async () => {
    const generated = {
      id: 1,
      title: "Note",
      body: "Body",
      createdAt: "created",
      updatedAt: "updated",
    };
    bindings.CreateNote.mockResolvedValue(generated);
    bindings.GetNote.mockResolvedValue(generated);
    bindings.UpdateNote.mockResolvedValue(undefined);
    bindings.DeleteNote.mockResolvedValue(undefined);

    await expect(createNote({ title: "Note", body: "Body" })).resolves.toMatchObject({
      id: 1,
      title: "Note",
    });
    await expect(getNote(1)).resolves.toMatchObject({ id: 1 });
    await expect(updateNote({ ...generated })).resolves.toBeUndefined();
    await expect(deleteNote(1)).resolves.toBeUndefined();

    expect(bindings.CreateNote).toHaveBeenCalledWith(
      expect.objectContaining({ source: expect.objectContaining({ title: "Note", body: "Body" }) }),
    );
    expect(bindings.GetNote).toHaveBeenCalledWith(1);
    expect(bindings.UpdateNote).toHaveBeenCalledOnce();
    expect(bindings.DeleteNote).toHaveBeenCalledWith(1);
  });

  it("propagates binding errors", async () => {
    bindings.ListNotes.mockRejectedValue(new Error("database unavailable"));

    await expect(listNotes()).rejects.toThrow("database unavailable");
  });
});
