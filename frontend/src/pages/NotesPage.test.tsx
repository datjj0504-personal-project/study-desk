import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { NotesPage } from "./NotesPage";

const { api } = vi.hoisted(() => ({
  api: {
    createNote: vi.fn(),
    deleteNote: vi.fn(),
    getNote: vi.fn(),
    listNotes: vi.fn(),
    updateNote: vi.fn(),
  },
}));

vi.mock("../services/notes", () => api);

afterEach(() => {
  cleanup();
  vi.resetAllMocks();
});

describe("NotesPage", () => {
  it("loads, creates, and selects a note", async () => {
    api.listNotes.mockResolvedValueOnce([]).mockResolvedValue([
      { id: 1, title: "Revision", body: "Chapter one", createdAt: "", updatedAt: "" },
    ]);
    api.createNote.mockResolvedValue({
      id: 1,
      title: "Revision",
      body: "Chapter one",
      createdAt: "",
      updatedAt: "",
    });

    render(<NotesPage />);
    expect(await screen.findByText("No saved notes yet.")).toBeTruthy();
    fireEvent.change(screen.getByLabelText("Title"), { target: { value: "Revision" } });
    fireEvent.change(screen.getByLabelText("Body"), { target: { value: "Chapter one" } });
    fireEvent.click(screen.getByRole("button", { name: "Save now" }));

    expect((await screen.findByRole("status")).textContent).toContain("Note created.");
    expect(api.createNote).toHaveBeenCalledWith({ title: "Revision", body: "Chapter one" });
    expect(screen.getByRole("button", { name: /Revision/ })).toBeTruthy();
  });

  it("updates and deletes a selected note", async () => {
    api.listNotes.mockResolvedValue([
      { id: 4, title: "Existing", body: "Old body", createdAt: "", updatedAt: "" },
    ]);
    api.updateNote.mockResolvedValue(undefined);
    api.getNote.mockResolvedValue({
      id: 4,
      title: "Existing",
      body: "New body",
      createdAt: "",
      updatedAt: "",
    });
    api.deleteNote.mockResolvedValue(undefined);

    render(<NotesPage />);
    fireEvent.change(await screen.findByLabelText("Body"), {
      target: { value: "New body" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Save now" }));
    await waitFor(() => expect(api.updateNote).toHaveBeenCalledWith(
      expect.objectContaining({ id: 4, title: "Existing", body: "New body" }),
    ));

    fireEvent.click(screen.getByRole("button", { name: "Delete" }));
    fireEvent.click(screen.getByRole("button", { name: "Confirm" }));
    await waitFor(() => expect(api.deleteNote).toHaveBeenCalledWith(4));
    expect(await screen.findByText("Note deleted.")).toBeTruthy();
  });

  it("surfaces loading errors", async () => {
    api.listNotes.mockRejectedValue(new Error("storage failed"));

    render(<NotesPage />);

    expect((await screen.findByRole("alert")).textContent).toContain("storage failed");
  });

  it("autosaves changes after the user pauses typing", async () => {
    api.listNotes.mockResolvedValue([
      { id: 5, title: "Draft", body: "Old", createdAt: "", updatedAt: "" },
    ]);
    api.updateNote.mockResolvedValue(undefined);
    api.getNote.mockResolvedValue({
      id: 5,
      title: "Draft",
      body: "New",
      createdAt: "",
      updatedAt: "",
    });

    render(<NotesPage />);
    fireEvent.change(await screen.findByLabelText("Body"), {
      target: { value: "New" },
    });

    expect(api.updateNote).not.toHaveBeenCalled();
    expect((await screen.findByRole("status")).textContent).toContain("Unsaved changes");
    await waitFor(() => expect(api.updateNote).toHaveBeenCalledOnce(), { timeout: 2000 });
    expect(await screen.findByText("Saved.")).toBeTruthy();
  });

  it("autosaves a new note after the title and body are entered", async () => {
    api.listNotes.mockResolvedValue([]);
    api.createNote.mockResolvedValue({
      id: 7,
      title: "New draft",
      body: "Saved automatically",
      createdAt: "",
      updatedAt: "",
    });

    render(<NotesPage />);
    fireEvent.change(await screen.findByLabelText("Title"), {
      target: { value: "New draft" },
    });
    fireEvent.change(screen.getByLabelText("Body"), {
      target: { value: "Saved automatically" },
    });

    await waitFor(() => expect(api.createNote).toHaveBeenCalledWith({
      title: "New draft",
      body: "Saved automatically",
    }), { timeout: 2000 });
    expect(await screen.findByText("Note created.")).toBeTruthy();
  });

  it("shows autosave failures and retries after another edit", async () => {
    api.listNotes.mockResolvedValue([
      { id: 6, title: "Draft", body: "Old", createdAt: "", updatedAt: "" },
    ]);
    api.updateNote
      .mockRejectedValueOnce(new Error("disk full"))
      .mockResolvedValueOnce(undefined);
    api.getNote.mockResolvedValue({
      id: 6,
      title: "Draft",
      body: "Retry",
      createdAt: "",
      updatedAt: "",
    });

    render(<NotesPage />);
    const editor = await screen.findByLabelText("Body");
    fireEvent.change(editor, { target: { value: "Failed draft" } });
    expect((await screen.findByRole("alert")).textContent).toContain("disk full");

    fireEvent.change(editor, { target: { value: "Retry" } });
    await waitFor(() => expect(api.updateNote).toHaveBeenCalledTimes(2), { timeout: 2000 });
    expect(await screen.findByText("Saved.")).toBeTruthy();
  });
});
