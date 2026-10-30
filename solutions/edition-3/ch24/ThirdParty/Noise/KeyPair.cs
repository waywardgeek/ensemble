using System;

namespace Noise
{
	/// <summary>
	/// A Diffie-Hellman private/public key pair.
	/// </summary>
	public sealed class KeyPair : IDisposable
	{
		private static readonly Curve25519 dh = new Curve25519();
		private readonly byte[] privateKey;
		private readonly byte[] publicKey;
		private bool disposed;

		/// <summary>
		/// Initializes a new instance of the <see cref="KeyPair"/> class.
		/// </summary>
		/// <param name="privateKey">The private key.</param>
		/// <param name="publicKey">The public key.</param>
		/// <exception cref="ArgumentNullException">
		/// Thrown if the <paramref name="privateKey"/> or the <paramref name="publicKey"/> is null.
		/// </exception>
		/// <exception cref="ArgumentException">
		/// Thrown if the lengths of the <paramref name="privateKey"/> or the <paramref name="publicKey"/> are invalid.
		/// </exception>
		internal KeyPair(byte[] privateKey, byte[] publicKey)
		{
			Exceptions.ThrowIfNull(privateKey, nameof(privateKey));
			Exceptions.ThrowIfNull(publicKey, nameof(publicKey));

			if (privateKey.Length != 32)
			{
				throw new ArgumentException("Private key must have length of 32 bytes.", nameof(privateKey));
			}

			if (publicKey.Length != 32)
			{
				throw new ArgumentException("Public key must have length of 32 bytes.", nameof(publicKey));
			}

			this.privateKey = privateKey;
			this.publicKey = publicKey;
		}

		/// <summary>
		/// Generates a new Diffie-Hellman key pair.
		/// </summary>
		/// <returns>A randomly generated private key and its corresponding public key.</returns>
		public static KeyPair Generate()
		{
			return dh.GenerateKeyPair();
		}

		/// <summary>
		/// Imports an X25519 private key and derives its public key using the
		/// existing DH implementation. The supplied bytes are copied.
		/// </summary>
		/// <param name="privateKey">Exactly 32 private key bytes.</param>
		/// <returns>An independently owned key pair.</returns>
		/// <exception cref="ArgumentException">The key is not 32 bytes.</exception>
		public static KeyPair Generate(ReadOnlySpan<byte> privateKey)
		{
			if (privateKey.Length != 32)
			{
				throw new ArgumentException("Private key must have length of 32 bytes.", nameof(privateKey));
			}

			return dh.GenerateKeyPair(privateKey);
		}

		/// <summary>
		/// Gets the private key.
		/// </summary>
		/// <exception cref="ObjectDisposedException">
		/// Thrown if the current instance has already been disposed.
		/// </exception>
		public byte[] PrivateKey
		{
			get
			{
				Exceptions.ThrowIfDisposed(disposed, nameof(KeyPair));
				return privateKey;
			}
		}

		/// <summary>
		/// Gets the public key.
		/// </summary>
		/// <exception cref="ObjectDisposedException">
		/// Thrown if the current instance has already been disposed.
		/// </exception>
		public byte[] PublicKey
		{
			get
			{
				Exceptions.ThrowIfDisposed(disposed, nameof(KeyPair));
				return publicKey;
			}
		}

		/// <summary>
		/// Erases the key pair from the memory.
		/// </summary>
		public void Dispose()
		{
			if (!disposed)
			{
				Utilities.ZeroMemory(privateKey);
				disposed = true;
			}
		}
	}
}
