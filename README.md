# gosyncmaildir - Maildir synchronization tool

This is a simple maildir synchronization tool. It might not do what you want it to do. It is
unrelated to syncmaildir, does not work yet, and probably needs a different name.


## Design

The design is based on git, but we don't keep old objects around, and we might
not keep all old trees.

### File layout

Both server and client store the following files:

- .gsmd/HEAD                    ASCII text file containing the ID of the last sync state. This
                                is effectively irrelevant on the server.
- .gsmd/<ID>                    A tree object. Tree objects are `encoding/gob` encoded lists of structs,
                                compressed with zstd.

A server should keep some old trees around to speed up syncing, whereas a client has no need to.

### Tree object

A tree is a mapping of IDs to filenames and mtimes. The ID of a file is currently the name of the
file without any flags or directory path. It is assumed to be unique throughout the maildir (recursively).

### Pull / Push sync

1. Server: Calculate DIFF(HEAD(CLIENT) -> TREE(SERVER))

    If the HEAD(CLIENT) is not present on the server, the server sends TREE(SERVER)
    instead of the difference, and the client has to calculate the difference.

1. Client: Calculate DIFF(HEAD -> TREE(CLIENT))

1. Client: Calculate the merged tree MERGED from HEAD and the two DIFFs, solving conflicts

    Conflicts will be resolved in some way, in general the server state wins.

    By option, deletions on the client should be ignored, and deleted files should be restored
    from the server, so that clients cannot accidentally delete emails.

1. Server: Apply(DIFF(MERGED -> TREE(CLIENT)))

    This involves sending new/modified emails to the server, deleting what the diff says needs
    deleting (if deleting is enabled), and moving files if the diff says they need moving.

1. Client: Apply(DIFF(MERGED -> TREE(SERVER)))

    This is basically the other way around.

### Transports

We should strive to provide transports over pipe, TCP, and HTTP. The transport could probably use `go/rpc`. All of these options seem easy to implement:

- Pipe just needs an `io.ReadWriteCloser` for the pipe ends. It is useful for transporting emails over SSH connections
- TCP and HTTP connections are implemented by the `go/rpc` directly.

We expect that most users are interested in the pipe transport, but the `http` transport, combined with an https frontend would enable syncing behind more restrictive firewalls.

Once we have calculated a difference, we can transport it as a stream starting with the gob-encoded tree difference object, followed by a stream of blobs referenced by their ids, for example (`id length data`).

### Server-side tree cache

The server needs to keep around some trees it send to clients, to speed up future syncs. If an old tree is not around, the client needs to fetch the entire current tree, rather than just the difference, and the tree might be relatively large.

It might make sense to introduce pack files to compress multiple trees into one object, or store some trees as deltas.

### Message IDs

Using a hash instead of the filename might provide a sensible improvement. It is suggested we use BLAKE2b hashes, as BLAKE2b is the fastest hash function that makes sense here. This avoids some pitfalls where we potentially could have the same ID twice. Compared to the Message-ID, it benefits us when receiving back emails we sent out - they'll have different headers, hence we can store them in both Sent/ and INBOX without issues.

That said, we should consider handling multiple links for a given ID.

### Merge algorithm

The merge algorithm needs some consideration. We generally follow a "server is right" kind of approach, but it might be sensible to adjust the merging of flags.

### Encryption

We want to support remotes (server or client) that store emails in an encrypted manner. For example, you might choose to GPG encrypt all emails on your server to have them encrypted at rest, but want them decrypted on your laptop.

Using file names as message ids, this is easy: The encrypted message on the server has the same name as its decrypted counterpart on the client, so we can easily match them. We can't compare sizes, though, which might be a useful optimization to determine if a file changed.

Assuming we derive the ID from the actual content, either by adding a size, or a hash of the file, we need to store those for encrypted and decrypted states. The server will only ever know the encrypted hash, the client has to know both.

Encryption is not reproducible, hence to map decrypted files back to encrypted files, whenever we decrypt a file, we need to store a mapping of the decrypted hash to the encrypted hash. If the file actually changed, we can then re-encrypt it, and update the encrypted tree and push it to the server.

Given that we probably do not also want to store the encrypted copy locally, but just submit it to the end point, it seems plausible that we put a place holder when calculating the IDs for the resulting tree (such as the decrypted hash), and later replace place holders with the actual IDs when writing the tree to the disk.

## Looping synchronization

We can loop the synchronization until there are no differences, ensuring that operations on either side while undergoing sync will be synced as well.