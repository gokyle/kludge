/*
 kludge provides a client interface to the kludge datastore.

 Clients call the Connect function to set up a connection to a
 datastore. This connection may be called from multiple goroutines
 safely. The Version method may be called to return the version
 of the datastore, and the ClientVersion function may be called to
 return the client's version.

 The datastore has three basic operations: Get, Set, and Del. These
 three operations are methods called on a DataStore value, and they
 return three values: a key value, a boolean indicating whether the
 key was present in the datastore before the operation was called,
 and an error value indicating any error that occurred.

 Additionally, the List method may be called to retrieve a list of all
 the keys present in the datastore.

 */
/*
   Copyright (c) 2013 Kyle Isom <kyle@gokyle.org>
   
   Permission to use, copy, modify, and distribute this software for any
   purpose with or without fee is hereby granted, provided that the above 
   copyright notice and this permission notice appear in all copies.
   
   THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
   WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
   MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
   ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
   WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
   ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
   OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE. 

 */
package kludge

