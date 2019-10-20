build:
	cd ffi && cargo build --release
	cp ffi/target/release/libevm_ffi.a libsputnikvm.a
	cp ffi/target/release/libevm_ffi.so libsputnikvm.so | true

build-musl:
	cd ffi && docker run --rm -it -v "$(shell pwd)/ffi":/home/rust/src ekidd/rust-musl-builder cargo build --release
	cp ffi/target/x86_64-unknown-linux-musl/release/libevm_ffi.a libsputnikvm.a

debug:
	cd ffi && cargo build
	cp ffi/target/debug/libevm_ffi.a libsputnikvm.a
	cp ffi/target/debug/libevm_ffi.so libsputnikvm.so | true

clean:
	cd ffi && cargo clean
	rm libsputnikvm.a libsputnikvm.so || true